package expr

import (
	"fmt"
	"slices"
	"strings"

	"github.com/samber/oops"
)

type Evaluator struct {
	ContextObject JSObject
	Functions     FunctionStore
}

func NewEvaluator(evalContext *EvalContext, funcs FunctionStore) (*Evaluator, error) {
	contextObject, err := jsObjectFromEvalContext(evalContext)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to create JSObject from EvalContext")
	}
	return &Evaluator{
		ContextObject: contextObject,
		Functions:     funcs,
	}, nil
}

func jsObjectFromEvalContext(evalContext *EvalContext) (JSObject, error) {
	jso, err := UnmarshalFromGo(*evalContext)
	if err != nil {
		return JSObject{}, err
	}
	if !jso.Object.IsPresent {
		return JSObject{}, oops.New("EvalContext did not contain an object")
	}

	return jso.Object.Value, nil
}

func (e *Evaluator) Evaluate(expression Node) (JSValue, error) {
	switch expr := expression.(type) {

	case *ObjectDerefNode, *ArrayDerefNode, *IndexAccessNode:
		receiverNode, jspath, err := e.evaluateAccessPath(expr)
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "failed to evaluate access path")
		}
		receiver, err := e.Evaluate(receiverNode)
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "failed to evaluate receiver of dereference")
		}
		evaluated, err := receiver.Access(jspath...)
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "failed to access path %s on %s", jspath.GoString(), receiver.GoString())
		}
		return evaluated, nil

	case *VariableNode:
		v, err := e.ContextObject.Access(JSPathSegment{String: Some(expr.Name)})
		if err != nil {
			return JSValue{}, err
		}
		return v, nil

	case *NullNode:
		return JSValue{Null: Some(struct{}{})}, nil

	case *BoolNode:
		return JSValue{Boolean: Some(expr.Value)}, nil

	case *IntNode:
		return JSValue{Int: Some(expr.Value)}, nil

	case *FloatNode:
		return JSValue{Float: Some(expr.Value)}, nil

	case *StringNode:
		return JSValue{String: Some(expr.Value)}, nil

	case *NotOpNode:
		child, err := e.Evaluate(expr.Operand)
		if err != nil {
			return JSValue{}, err
		}

		return JSValue{Boolean: Some(!child.toBool())}, nil

	case *CompareOpNode:
		left, err := e.Evaluate(expr.Left)
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "failed to evaluate left operand of %s at %s (%s)", expr.Kind.String(), expr.Left.Token(), expr.Left.Token().Value)
		}
		right, err := e.Evaluate(expr.Right)
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "failed to evaluate right operand of %s at %s (%s)", expr.Kind.String(), expr.Right.Token(), expr.Right.Token().Value)
		}
		var compareResult float64
		switch {
		case left.canNumber() && right.canNumber():
			compareResult = left.number() - right.number()
		case left.String.IsPresent && right.String.IsPresent:
			compareResult = float64(strings.Compare(left.String.Value, right.String.Value))
		default:
			return JSValue{}, oops.Errorf("left and right side of %s have incompatible types for comparison: '%s' '%s'", expr.Kind.String(), left.Type(), right.Type())
		}

		switch expr.Kind {
		case CompareOpNodeKindInvalid:
			return JSValue{}, oops.Errorf("invalid comparison operator: %s", expr.Kind.String())
		case CompareOpNodeKindLess:
			return JSValue{Boolean: Some(compareResult < 0)}, nil
		case CompareOpNodeKindLessEq:
			return JSValue{Boolean: Some(compareResult <= 0)}, nil
		case CompareOpNodeKindGreater:
			return JSValue{Boolean: Some(compareResult > 0)}, nil
		case CompareOpNodeKindGreaterEq:
			return JSValue{Boolean: Some(compareResult >= 0)}, nil
		case CompareOpNodeKindEq:
			return JSValue{Boolean: Some(compareResult == 0)}, nil
		case CompareOpNodeKindNotEq:
			return JSValue{Boolean: Some(compareResult != 0)}, nil
		default:
			panic(fmt.Sprintf("value of CompareOpNodeKind(%d) is not part of enum", expr.Kind))
		}

	case *LogicalOpNode:

		switch expr.Kind {
		case LogicalOpNodeKindAnd:
			left, err := e.Evaluate(expr.Left)
			if err != nil {
				return JSValue{}, oops.Wrapf(err, "failed to evaluate left operand of %s at %s (%s)", expr.Kind.String(), expr.Left.Token(), expr.Left.Token().Value)
			}
			if !left.toBool() {
				return left, nil
			}
			right, err := e.Evaluate(expr.Right)
			if err != nil {
				return JSValue{}, oops.Wrapf(err, "failed to evaluate right operand of %s at %s (%s)", expr.Kind.String(), expr.Right.Token(), expr.Right.Token().Value)
			}
			return right, nil
		case LogicalOpNodeKindOr:
			left, err := e.Evaluate(expr.Left)
			if err != nil {
				return JSValue{}, oops.Wrapf(err, "failed to evaluate left operand of %s at %s (%s)", expr.Kind.String(), expr.Left.Token(), expr.Left.Token().Value)
			}
			if left.toBool() {
				return left, nil
			}
			right, err := e.Evaluate(expr.Right)
			if err != nil {
				return JSValue{}, oops.Wrapf(err, "failed to evaluate right operand of %s at %s (%s)", expr.Kind.String(), expr.Right.Token(), expr.Right.Token().Value)
			}
			return right, nil
		default:
			return JSValue{}, oops.Errorf("unexpected logical operator at %s", expr.Token().String())
		}

	case *FuncCallNode:
		return e.evaluateFunctionCall(expr)

	default:
		return JSValue{}, oops.Errorf("unexpected expression type %T, token %s (%s)", expr, expr.Token().String(), expr.Token().Value)
	}
}

func (e *Evaluator) evaluateAccessPath(expression Node) (root Node, p JSPath, er error) {
	var resultReversed []JSPathSegment
	currentExpr := expression
	const maxDepth = 1_000_000

	for i := 0; i < maxDepth; i++ {
		switch expr := currentExpr.(type) {
		case *IndexAccessNode:
			evaled, err := e.Evaluate(expr.Index)
			if err != nil {
				return nil, nil, oops.Wrapf(err, "failed to evaluate index expression %s", expr.Token().String())
			}
			switch {
			case evaled.Int.IsPresent:
				resultReversed = append(resultReversed, JSPathSegment{Int: evaled.Int})
			case evaled.String.IsPresent:
				resultReversed = append(resultReversed, JSPathSegment{String: evaled.String})
			default:
				return nil, nil, oops.Errorf("trying to index access using incompatible type %T, token %s (%s)", evaled.Type, expr.Token().String(), expr.Token().Value)
			}
			currentExpr = expr.Operand
		case *ObjectDerefNode:
			resultReversed = append(resultReversed, JSPathSegment{String: Some(expr.Property)})
			currentExpr = expr.Receiver
		case *ArrayDerefNode:
			resultReversed = append(resultReversed, JSPathSegment{Star: Some(struct{}{})})
			currentExpr = expr.Receiver
		default:
			slices.Reverse(resultReversed)
			return currentExpr, JSPath(resultReversed), nil
		}
	}
	return nil, nil, oops.Errorf("potential infinite recursion detected while evaluating access path at %s", expression.Token())
}

func (e *Evaluator) evaluateFunctionCall(expr *FuncCallNode) (JSValue, error) {
	args := make([]JSValue, len(expr.Args))
	for i, arg := range expr.Args {
		value, err := e.Evaluate(arg)
		if err != nil {
			return JSValue{}, oops.Wrapf(err, "failed to evaluate argument %d", i)
		}
		args[i] = value
	}

	fn, ok := e.Functions[strings.ToLower(expr.Callee)]
	if !ok {
		return JSValue{}, oops.Errorf("function %s not found", expr.Callee)
	}

	v, err := fn(args...)
	if err != nil {
		argsTypes := make([]string, len(args))
		for i, arg := range args {
			argsTypes[i] = string(arg.Type())
		}
		return JSValue{}, oops.Wrapf(err, "error from function %s with args (%s)", expr.Callee, strings.Join(argsTypes, ", "))
	}

	return v, nil
}
