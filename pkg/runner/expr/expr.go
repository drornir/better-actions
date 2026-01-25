// Package expr provides an expression parser and evaluator, usually surrounded by `${{ }}` .
// Paring is based on https://github.com/rhysd/actionlint, meaning a lot of copy pasting with extensions
// to allow more than the basic functionality.
package expr

import (
	"fmt"
	"strings"

	"github.com/samber/oops"
)

// TODO
/*

## The Insert Directive

This is an **undocumented GitHub Actions feature** (as the comment on line 270 mentions) that allows you to **merge/flatten a nested map into its parent map**.

### How it works:

When you have a YAML key that matches the pattern `${{ insert }}`, the code:

1. **Evaluates the value** associated with that key (which must be a mapping/object)
2. **Merges all key-value pairs** from that nested map directly into the parent map
3. **Removes the `${{ insert }}` key itself** — it's just a directive, not a real key

### Example Usage:

```/dev/null/example.yaml#L1-10
env:
  STATIC_VAR: "hello"
  ${{ insert }}: ${{ fromJSON('{"DYNAMIC_VAR1": "value1", "DYNAMIC_VAR2": "value2"}') }}
  ANOTHER_VAR: "world"
```

This would effectively become:

```/dev/null/result.yaml#L1-5
env:
  STATIC_VAR: "hello"
  DYNAMIC_VAR1: "value1"
  DYNAMIC_VAR2: "value2"
  ANOTHER_VAR: "world"
```

### Key code logic (lines 297-304):

```act/pkg/runner/expression.go#L297-304
		if k.Decode(&sk) == nil && insertDirective.MatchString(sk) {
			if ev.Kind != yaml.MappingNode {
				return nil, fmt.Errorf("failed to insert node %v into mapping %v unexpected type %v expected MappingNode", ev, node, ev.Kind)
			}
			if err := changed(); err != nil {
				return nil, err
			}
			ret.Content = append(ret.Content, ev.Content...)
```

The value **must be a MappingNode** (a YAML map), and its contents (`ev.Content...`) are spread/merged directly into the parent map's content.

### Use Case:

This is useful when you want to dynamically inject multiple environment variables or other map entries from an expression like `fromJSON()`, `toJSON()`, or a context variable that contains an object.


---------------------------------------------------------------------------------------------------------------



## Automatic Array Flattening

Look at lines 344-348:

```act/pkg/runner/expression.go#L344-348
			// GitHub has this undocumented feature to merge sequences / arrays
			// We have a nested sequence via evaluation, merge the arrays
			if ev.Kind == yaml.SequenceNode && !wasseq {
				ret.Content = append(ret.Content, ev.Content...)
			} else {
```

### How it works:

When an expression **evaluates to a sequence** (array), and the original value **wasn't already a sequence** in the YAML, the array gets **flattened/spread** into the parent array.

The key condition is `!wasseq` — it only flattens if the original YAML node wasn't already a sequence (to preserve intentionally nested arrays).

### Example:

```/dev/null/example.yaml#L1-5
steps:
  - name: "First step"
    run: echo "hello"
  - ${{ fromJSON('["item1", "item2", "item3"]') }}
  - name: "Last step"
    run: echo "goodbye"
```

The `fromJSON` expression returns an array, and since the original YAML value was a scalar (the expression string), not a sequence, it gets flattened:

```/dev/null/result.yaml#L1-7
steps:
  - name: "First step"
    run: echo "hello"
  - item1
  - item2
  - item3
  - name: "Last step"
    run: echo "goodbye"
```

### Key Difference from Insert Directive:

| Feature | Maps (insert) | Arrays |
|---------|---------------|--------|
| Syntax | Explicit `${{ insert }}` key required | Automatic — no special syntax |
| Trigger | Key matches the insert pattern | Expression evaluates to an array |
| Condition | N/A | Only if original wasn't already a sequence (`!wasseq`) |

So for arrays, it's **implicit** — just use an expression that returns an array, and it auto-flattens. For maps, you need the **explicit `${{ insert }}`** key.

## User

give me examples of when wasseq is true and false

## Assistant

<thinking>
The user wants to understand the `wasseq` variable better with concrete examples. Let me think about this:

- `wasseq` is set to `true` when `v.Kind == yaml.SequenceNode` - meaning the original YAML node was already a sequence/array before evaluation
- `wasseq` is `false` when the original YAML node was something else (like a scalar string containing an expression)

This distinction determines whether the result gets flattened or kept nested.
</thinking>

Here are concrete examples showing both cases:

## `wasseq = false` (Flattens)

The original YAML value is a **scalar** (string containing an expression), not a sequence:

```/dev/null/example1.yaml#L1-4
items:
  - static_item
  - ${{ fromJSON('["dynamic1", "dynamic2"]') }}
  - another_static
```

The `${{ fromJSON(...) }}` is a **scalar string** in the YAML → `wasseq = false`

**Result:** Array gets flattened:
```/dev/null/result1.yaml#L1-5
items:
  - static_item
  - dynamic1
  - dynamic2
  - another_static
```

---

## `wasseq = true` (Preserves nesting)

The original YAML value is **already a sequence**:

```/dev/null/example2.yaml#L1-5
items:
  - static_item
  - - ${{ variables.item1 }}
    - ${{ variables.item2 }}
  - another_static
```

The middle item is a **nested YAML array** (sequence node) → `wasseq = true`

**Result:** Nesting is preserved:
```/dev/null/result2.yaml#L1-5
items:
  - static_item
  - - resolved_item1
    - resolved_item2
  - another_static
```

---

## Side-by-side comparison

| YAML Source | `wasseq` | Behavior |
|-------------|----------|----------|
| `- ${{ fromJSON('["a","b"]') }}` | `false` (scalar) | Flattens → `- a` `- b` |
| `- - ${{ x }}` | `true` (sequence) | Preserves → `- - value` |

---

## Why this matters

This lets you **intentionally create nested arrays** when you need them:

```/dev/null/matrix-example.yaml#L1-6
# Flattened (wasseq=false) - probably what you want for dynamic items
include: ${{ fromJSON('[{"os":"linux"},{"os":"mac"}]') }}

# Nested (wasseq=true) - intentionally nested structure
matrix:
  - - a
    - b
```

The `!wasseq` check prevents the code from accidentally flattening arrays that you explicitly wrote as nested in the YAML.
*/

type Evaluator struct {
	ll *LowLevelEvaluator
}

func NewEvaluator(evalContext *EvalContext) (*Evaluator, error) {
	ll, err := NewLowLevelEvaluator(evalContext, DefaultFunctions)
	if err != nil {
		return nil, err
	}
	return &Evaluator{
		ll: ll,
	}, nil
}

// EvaluateTemplate evaluates a string that might contain expressions surrounded by `${{ }}`
func (e *Evaluator) EvaluateTemplate(template string) (string, error) {
	const dollarDollar = "\x00DOLOAR_DOLLAR\x00"
	escapedTemplate := strings.ReplaceAll(template, "$$", dollarDollar)
	if !strings.Contains(template, "${{") {
		return template, nil
	}

	asRunes := []rune(escapedTemplate)
	result := strings.Builder{}
	for {
		openingIdx := strings.Index(string(asRunes), "${{")
		if openingIdx == -1 {
			result.WriteString(string(asRunes))
			break
		}
		asRunes = asRunes[openingIdx+len("${{"):]
		closingIdx := strings.Index(string(asRunes), "}}")
		if closingIdx == -1 {
			return "", oops.Errorf("can't find closing braces for ${{%s", string(asRunes))
		}
		// note the parser expects the closing }}
		expr, rest := string(asRunes[:closingIdx+len("}}")]), asRunes[closingIdx+len("}}"):]
		asRunes = rest

		expr = strings.ReplaceAll(expr, dollarDollar, "$$")
		parsed, perr := NewParser().Parse(NewExprLexer(expr))
		if perr != nil {
			return "", oops.Wrapf(perr, "parsing expression ${{%s", expr)
		}
		evaled, err := e.ll.Evaluate(parsed)
		if err != nil {
			return "", oops.Wrapf(err, "evaluating expression ${{%s", expr)
		}
		coerced, err := castToString(evaled)
		if err != nil {
			return "", oops.Wrapf(err, "coercing expression to string ${{%s", expr)
		}
		result.WriteString(coerced)
	}

	final := strings.ReplaceAll(result.String(), dollarDollar, "$$")
	return final, nil
}

// EvaluateExpression evaluates a string that might be an expression or a template. Used in e.g 'if'.
func (e *Evaluator) EvaluateExpression(expressionOrTemplate string) (string, error) {
	if !strings.HasPrefix(strings.TrimSpace(expressionOrTemplate), "${{") {
		expressionOrTemplate = fmt.Sprintf("${{ %s }}", expressionOrTemplate)
	}

	return e.EvaluateTemplate(expressionOrTemplate)
}
