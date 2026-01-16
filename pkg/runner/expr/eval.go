package expr

import "github.com/samber/oops"

type Evaluator struct {
	ContextObject JSObject
}

func NewEvaluator(evalContext *EvalContext) (*Evaluator, error) {
	contextObject, err := jsObjectFromEvalContext(evalContext)
	if err != nil {
		return nil, oops.Wrapf(err, "failed to create JSObject from EvalContext")
	}
	return &Evaluator{
		ContextObject: contextObject,
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

	case *VariableNode:
	case *NullNode:
	case *BoolNode:
	case *IntNode:
	case *FloatNode:
	case *StringNode:
	case ObjectDerefNode:
	case ArrayDerefNode:
	case *IndexAccessNode:
	case *NotOpNode:
	case *CompareOpNode:
	case *LogicalOpNode:
	case *FuncCallNode:

	default:
		_ = expr
		return JSValue{}, oops.Errorf("unexpected expression type %s", expression.Token().String())
	}

	panic("")
}
