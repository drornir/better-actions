package expr

import (
	"fmt"
	"strings"

	"github.com/samber/oops"
)

type (
	JSObject map[string]JSValue
	JSArray  []JSValue

	JSValue struct {
		Object    Maybe[JSObject]
		Array     Maybe[JSArray]
		String    Maybe[string]
		Number    Maybe[float64]
		Boolean   Maybe[bool]
		Null      Maybe[struct{}]
		Undefined Maybe[struct{}]
	}
)

type Maybe[T any] struct {
	Value     T
	IsPresent bool
}

func Some[T any](value T) Maybe[T] {
	return Maybe[T]{Value: value, IsPresent: true}
}

func (j JSValue) GoValue() any {
	switch {
	case j.Object.IsPresent:
		return j.Object.Value.GoValue()
	case j.Array.IsPresent:
		return j.Array.Value.GoValue()
	case j.String.IsPresent:
		return j.String.Value
	case j.Number.IsPresent:
		return j.Number.Value
	case j.Boolean.IsPresent:
		return j.Boolean.Value
	case j.Null.IsPresent:
		return nil
	case j.Undefined.IsPresent:
		return nil
	default:
		return nil
	}
}

func (j JSArray) GoValue() []any {
	gv := make([]any, 0, len(j))
	for _, item := range j {
		gv = append(gv, item.GoValue())
	}
	return gv
}

func (j JSObject) GoValue() map[string]any {
	gv := make(map[string]any)
	for key, value := range j {
		gv[key] = value.GoValue()
	}
	return gv
}

type JSPathSegment struct {
	String Maybe[string]
	Int    Maybe[int]
	Star   Maybe[struct{}]
}

type ErrJSAccess struct {
	Type    string
	Segment JSPathSegment
}

func (e ErrJSAccess) Error() string {
	return fmt.Sprintf("TypeError: Cannot read properties of %s (reading %#v)", e.Type, e.Segment)
}

func (j JSValue) Access(p ...JSPathSegment) (JSValue, error) {
	if len(p) == 0 {
		return j, nil
	}
	switch {
	case j.Object.IsPresent:
		return j.Object.Value.Access(p...)
	case j.Array.IsPresent:
		return j.Array.Value.Access(p...)
	case j.String.IsPresent:
		if !p[0].Int.IsPresent {
			return JSValue{}, ErrJSAccess{Type: "string", Segment: p[0]}
		}
		index := p[0].Int.Value
		if index < 0 || index >= len([]rune(j.String.Value)) {
			return JSValue{}, ErrJSAccess{Type: "string", Segment: p[0]}
		}

		asJSArray := make(JSArray, 0, len(j.String.Value))
		for char := range strings.SplitSeq(j.String.Value, "") {
			asJSArray = append(asJSArray, JSValue{String: Some(char)})
		}
		return asJSArray.Access(p...)
	case j.Number.IsPresent:
		return JSValue{Number: Some(j.Number.Value)}, nil
	case j.Boolean.IsPresent:
		return JSValue{Boolean: Some(j.Boolean.Value)}, nil
	case j.Null.IsPresent:
		return JSValue{Null: Some(struct{}{})}, nil
	case j.Undefined.IsPresent:
		return JSValue{Undefined: Some(struct{}{})}, nil
	default:
		return JSValue{}, ErrJSAccess{Type: "undefined", Segment: p[0]}
	}
}

func (j JSObject) Access(p ...JSPathSegment) (JSValue, error) {
	if len(p) == 0 {
		return JSValue{Object: Some(j)}, nil
	}
	if j == nil {
		return JSValue{}, ErrJSAccess{Type: "undefined", Segment: p[0]}
	}

	if p[0].Star.IsPresent {
		collection := j.values()
		output := make([]JSValue, 0, len(collection))
		for _, item := range collection {
			mapped, err := item.Access(p[1:]...)
			if err != nil {
				return JSValue{}, oops.Wrapf(err, "error in star expansion of %#v", item)
			}
			output = append(output, mapped)
		}
		return JSValue{Array: Some(JSArray(output))}, nil
	}

	if p[0].String.IsPresent {
		value, ok := j[p[0].String.Value]
		if !ok {
			return JSValue{Undefined: Some(struct{}{})}, nil
		}
		return value.Access(p[1:]...)
	}

	if p[0].Int.IsPresent {
		index := p[0].Int.Value
		key := fmt.Sprintf("%d", index)
		value, ok := j[key]
		if !ok {
			return JSValue{Undefined: Some(struct{}{})}, nil
		}
		return value.Access(p[1:]...)
	}

	return JSValue{}, ErrJSAccess{Type: "unexpected", Segment: p[0]}
}

func (j JSObject) values() []JSValue {
	values := make([]JSValue, 0, len(j))
	for _, value := range j {
		values = append(values, value)
	}
	return values
}

func (j JSArray) Access(p ...JSPathSegment) (JSValue, error) {
	if len(p) == 0 {
		return JSValue{Array: Some(j)}, nil
	}
	if j == nil {
		return JSValue{}, ErrJSAccess{Type: "undefined", Segment: p[0]}
	}

	if p[0].Star.IsPresent {
		output := make(JSArray, 0, len(j))
		for _, item := range j {
			mapped, err := item.Access(p[1:]...)
			if err != nil {
				return JSValue{}, oops.Wrapf(err, "error in star expansion of %#v", item)
			}
			output = append(output, mapped)
		}
		return JSValue{Array: Some(output)}, nil
	}

	if p[0].Int.IsPresent {
		index := p[0].Int.Value
		if index < 0 || index >= len(j) {
			return JSValue{Undefined: Some(struct{}{})}, nil
		}
		return j[index].Access(p[1:]...)
	}

	return JSValue{}, ErrJSAccess{Type: "unexpected", Segment: p[0]}
}

// debugging helpers

func (j JSValue) GoString() string {
	switch {
	case j.Object.IsPresent:
		return j.Object.Value.GoString()
	case j.Array.IsPresent:
		return j.Array.Value.GoString()
	case j.String.IsPresent:
		return fmt.Sprintf("%q", j.String.Value)
	case j.Number.IsPresent:
		return fmt.Sprintf("%f", j.Number.Value)
	case j.Boolean.IsPresent:
		return fmt.Sprintf("%t", j.Boolean.Value)
	case j.Null.IsPresent:
		return "null"
	case j.Undefined.IsPresent:
		return "undefined"
	default:
		return fmt.Sprintf("%v", j)
	}
}

func (j JSObject) GoString() string {
	sb := strings.Builder{}
	sb.WriteString("{ ")
	for k, v := range j {
		sb.WriteString(k)
		sb.WriteString(":")
		sb.WriteString(fmt.Sprintf("%#v", v))
		sb.WriteString(", ")
	}
	sb.WriteString(" }")
	return sb.String()
}

func (j JSArray) GoString() string {
	sb := strings.Builder{}
	sb.WriteString("[ ")
	for _, v := range j {
		sb.WriteString(fmt.Sprintf("%#v", v))
		sb.WriteString(", ")
	}
	sb.WriteString(" ]")
	return sb.String()
}

func (p JSPathSegment) GoString() string {
	if p.Int.IsPresent {
		return fmt.Sprintf("%d", p.Int.Value)
	}
	if p.String.IsPresent {
		return fmt.Sprintf("%q", p.String.Value)
	}
	if p.Star.IsPresent {
		return "*"
	}
	return fmt.Sprintf("%v", p)
}
