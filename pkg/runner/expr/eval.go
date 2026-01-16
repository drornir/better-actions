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
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *NullNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *BoolNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *IntNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *FloatNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *StringNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case ObjectDerefNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case ArrayDerefNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *IndexAccessNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *NotOpNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *CompareOpNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *LogicalOpNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)
	case *FuncCallNode:
		return JSValue{}, oops.Errorf("TODO: %T unimplemented", expression)

	default:
		_ = expr
		return JSValue{}, oops.Errorf("unexpected expression type %s", expression.Token().String())
	}
}
