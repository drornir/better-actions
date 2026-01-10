package expr

type Evaluator struct {
	OriginalContext *EvalContext
	ContextObject   JSObject
}

func NewEvaluator(evalContext *EvalContext) *Evaluator {
	return &Evaluator{
		OriginalContext: evalContext,
	}
}

func (e *Evaluator) Evaluate(expression Node) (string, error) {
	panic("not implemented")
}
