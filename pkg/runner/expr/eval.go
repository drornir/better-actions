package expr

import (
	"cmp"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/samber/oops"
)

type LowLevelEvaluator struct {
	ContextObject JSObject
	Functions     FunctionStore
}

func NewLowLevelEvaluator(evalContext *EvalContext, funcs FunctionStore) (*LowLevelEvaluator, error) {
	contextObject, err := jsObjectFromEvalContext(evalContext)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to parse EvalContext")
	}
	return &LowLevelEvaluator{
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

func (e *LowLevelEvaluator) Evaluate(expression Node) (JSValue, error) {
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

		compareResult, ok := compareJSValues(left, right)
		if !ok {
			return JSValue{Boolean: Some(false)}, nil
		}

		switch expr.Kind {
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

func (e *LowLevelEvaluator) evaluateAccessPath(expression Node) (root Node, p JSPath, er error) {
	var resultReversed []JSPathSegment
	currentExpr := expression
	const maxDepth = 1_000_000

	for range maxDepth {
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

func (e *LowLevelEvaluator) evaluateFunctionCall(expr *FuncCallNode) (JSValue, error) {
	if slices.Contains([]string{"success", "always", "cancelled", "failure"}, strings.ToLower(expr.Callee)) {
		return e.evaluateStatusFunction(expr.Callee)
	}

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

func (e *LowLevelEvaluator) evaluateStatusFunction(callee string) (JSValue, error) {
	panic("unimplemented")
}

// compareJSValues compares two JSValues and returns an integer indicating their order.
// the boolean is set to false when a NaN was part of the computation
// spec: https://docs.github.com/en/actions/reference/workflows-and-actions/expressions#operators
// BUG: in the spec, objects and arrays should be equal according to their address but it's not working here.
// Since this is a niche behavior, I'm not fixing it until it becomes a real issue.
func compareJSValues(a, b JSValue) (int, bool) {
	switch {
	case a.canNumber() && b.canNumber():
		return compareNumberJSValues(a.number(), b.number()), true

	case a.Type() == b.Type():
		switch {
		case a.String.IsPresent:
			return strings.Compare(a.String.Value, b.String.Value), true
		case a.Null.IsPresent, a.Undefined.IsPresent:
			return 0, true
		case a.Array.IsPresent:
			// this is horrible but is compatibility with github
			// has a bug
			eq := len(a.Array.Value) == len(b.Array.Value) && (len(a.Array.Value) == 0 || &(a.Array.Value) == &(b.Array.Value))
			if eq {
				return 0, true
			} else {
				return 0, false
			}
		case a.Object.IsPresent:
			// this is horrible but is compatibility with github
			// has a bug
			eq := len(a.Object.Value) == len(b.Object.Value) && (len(a.Object.Value) == 0 || &(a.Object.Value) == &(b.Object.Value))
			if eq {
				return 0, true
			} else {
				return 0, false
			}
		case a.Boolean.IsPresent:
			if a.Boolean.Value == b.Boolean.Value {
				return 0, true
			} else {
				return 1, true
			}
		default:
			panic("unreachable")
		}

	default:
		ac, aok := coerceJSValueToNumber(a)
		bc, bok := coerceJSValueToNumber(b)
		if !aok || !bok {
			return 0, false
		}
		return compareNumberJSValues(ac, bc), true
	}
}

// coerceJSValueToNumber converts a JSValue to a number.
// If the value is NaN, returns false
func coerceJSValueToNumber(v JSValue) (float64, bool) {
	switch {
	case v.canNumber():
		return v.number(), true
	case v.String.IsPresent:
		if v.String.Value == "" {
			return 0, true
		}
		if n, err := strconv.ParseFloat(v.String.Value, 64); err == nil {
			return n, true
		} else {
			var nerr *strconv.NumError
			if errors.As(err, &nerr) && errors.Is(nerr.Err, strconv.ErrRange) && math.IsInf(n, 0) {
				return n, true
			}
			return 0, false
		}
	case v.Boolean.IsPresent:
		if v.Boolean.Value {
			return 1, true
		}
		return 0, true
	case v.Null.IsPresent || v.Undefined.IsPresent:
		return 0, true
	default:
		return 0, false
	}
}

func compareNumberJSValues[T cmp.Ordered](a, b T) int {
	if a < b {
		return -1
	} else if a > b {
		return 1
	}
	return 0
}
