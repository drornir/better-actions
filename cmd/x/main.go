package main

import (
	"fmt"

	"github.com/drornir/better-actions/pkg/runner/expr"
)

func main() {
	exprs := []string{
		"github",
		"github.event",
		"arr[0]",
		"arr.*",
		"true || false",
		"true && false",
		"(true && false) || x",
		"true && false || x",
		"true || false && x",
		"true || (false && x)",
		"!true",
		"true || (!false && x)",
		"42 > 24",
		"foo(42, 43, 44)",
	}

	for _, ex := range exprs {
		fmt.Println(ex)
		fmt.Println("-------")

		// lexed, _, err := expr.LexExpression(ex + "}}")
		// if err != nil {
		// 	fmt.Printf("error lexing expression ${{ %s }}': %s\n", ex, err)
		// 	continue
		// }

		parsed, perr := expr.NewParser().Parse(expr.NewExprLexer(ex + "}}"))
		if perr != nil {
			fmt.Printf("error parsing expression ${{ %s }}': %s\n", ex, perr)
			continue
		}

		fmt.Println(expr.VisualizeAST(parsed))

		fmt.Println()
	}
}
