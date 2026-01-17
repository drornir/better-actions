package expr_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/drornir/better-actions/pkg/runner/expr"
)

func TestEvaluate(t *testing.T) {
	testCases := []struct {
		expr     string
		expected string
	}{
		{"true", "true"},
		{"!true", "false"},
		{"true || false", "true"},
		{"true && false", "false"},
		{"true && false || false", "false"},
		{"true && false || true", "true"},
		{"true || false && true", "true"},
		{"42 > 24", "true"},
		{"42 >= 24", "true"},
		{"42 < 24", "false"},
		{"42 <= 24", "false"},
		{"42", "42"},
		{"'hello'", "\"hello\""},
		{"null", "null"},
	}

	for _, tc := range testCases {
		t.Run(tc.expr, func(t *testing.T) {
			evaluator, err := expr.NewEvaluator(&expr.EvalContext{}, expr.DefaultFunctions)
			if !assert.NoErrorf(t, err, "initializing evaluator") {
				return
			}

			ast, parseErr := expr.NewParser().Parse(expr.NewExprLexer(tc.expr + "}}"))
			var errFix error
			if parseErr != nil {
				errFix = parseErr
			}
			if !assert.NoErrorf(t, errFix, "parsing expression") {
				return
			}

			result, err := evaluator.Evaluate(ast)
			if !assert.NoErrorf(t, err, "evaluating expression") {
				return
			}

			resAsJSON, err := result.MarshalJSON()
			if !assert.NoErrorf(t, err, "marshaling result") {
				return
			}

			assert.JSONEq(t, tc.expected, string(resAsJSON))
		})
	}
}
