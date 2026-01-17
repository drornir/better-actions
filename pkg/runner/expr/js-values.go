package expr

import (
	"encoding/json/v2"
	"fmt"
	"reflect"
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
		Float     Maybe[float64]
		Int       Maybe[int64]
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

type JSValueType string

const (
	ObjectType    JSValueType = "object"
	ArrayType     JSValueType = "array"
	StringType    JSValueType = "string"
	NumberType    JSValueType = "number"
	IntType       JSValueType = "int"
	BooleanType   JSValueType = "boolean"
	NullType      JSValueType = "null"
	UndefinedType JSValueType = "undefined"
)

func (j JSValue) Type() JSValueType {
	switch {
	case j.Object.IsPresent:
		return ObjectType
	case j.Array.IsPresent:
		return ArrayType
	case j.String.IsPresent:
		return StringType
	case j.Float.IsPresent:
		return NumberType
	case j.Int.IsPresent:
		return IntType
	case j.Boolean.IsPresent:
		return BooleanType
	case j.Null.IsPresent:
		return NullType
	case j.Undefined.IsPresent:
		return UndefinedType
	default:
		panic("unreachable")
	}
}

func (j JSValue) GoValue() any {
	switch {
	case j.Object.IsPresent:
		return j.Object.Value.GoValue()
	case j.Array.IsPresent:
		return j.Array.Value.GoValue()
	case j.String.IsPresent:
		return j.String.Value
	case j.Float.IsPresent:
		return j.Float.Value
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

func (j JSValue) toBool() bool {
	return (j.Boolean.IsPresent && j.Boolean.Value == true) ||
		j.Object.IsPresent ||
		j.Array.IsPresent ||
		(j.Float.IsPresent && j.Float.Value != 0) ||
		(j.Int.IsPresent && j.Int.Value != 0) ||
		(j.String.IsPresent && j.String.Value != "")
}

func (j JSValue) canNumber() bool {
	return j.Float.IsPresent || j.Int.IsPresent
}

func (j JSValue) number() float64 {
	if j.Float.IsPresent {
		return j.Float.Value
	}
	if j.Int.IsPresent {
		return float64(j.Int.Value)
	}
	panic(fmt.Sprintf("JSValue is not a number: is %s", j.Type()))
}

type JSPathSegment struct {
	String Maybe[string]
	Int    Maybe[int64]
	Star   Maybe[struct{}]
}

type JSPath []JSPathSegment

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
		if index < 0 || index >= int64(len([]rune(j.String.Value))) {
			return JSValue{}, ErrJSAccess{Type: "string", Segment: p[0]}
		}

		asJSArray := make(JSArray, 0, len(j.String.Value))
		for char := range strings.SplitSeq(j.String.Value, "") {
			asJSArray = append(asJSArray, JSValue{String: Some(char)})
		}
		return asJSArray.Access(p...)
	default:
		return JSValue{}, ErrJSAccess{Type: string(j.Type()), Segment: p[0]}
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
		if index < 0 || index >= int64(len(j)) {
			return JSValue{Undefined: Some(struct{}{})}, nil
		}
		return j[index].Access(p[1:]...)
	}

	return JSValue{}, ErrJSAccess{Type: "unexpected", Segment: p[0]}
}

// unmarshal

type ErrJSUnmarshal struct {
	Msg   string
	Value any
}

func (e ErrJSUnmarshal) Error() string {
	return fmt.Sprintf("error unmarshalling %#v: %s", e.Value, e.Msg)
}

func UnmarshalFromGo(value any) (JSValue, error) {
	v := JSValue{}
	if err := v.UnmarshalFromGo(value); err != nil {
		return JSValue{}, err
	}
	return v, nil
}

func (j *JSValue) UnmarshalFromGo(value any) error {
	if value == nil {
		j.Null = Some(struct{}{})
		return nil
	}

	switch value := value.(type) {
	case JSValue:
		*j = value
		return nil
	case JSObject:
		j.Object = Some(value)
		return nil
	case JSArray:
		j.Array = Some(value)
		return nil
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	default:
		return ErrJSUnmarshal{Msg: "unexpected type", Value: value}
	case reflect.Struct:
		asJSONMap := make(map[string]any, rv.NumField())
		structAsJSONString, err := json.Marshal(value)
		if err != nil {
			return ErrJSUnmarshal{Msg: "error marshalling struct", Value: value}
		}
		if err := json.Unmarshal(structAsJSONString, &asJSONMap); err != nil {
			return ErrJSUnmarshal{Msg: "error unmarshalling struct", Value: value}
		}
		return j.UnmarshalFromGo(asJSONMap)

	case reflect.Map:
		o := JSObject{}

		mapkeys := rv.MapKeys()
		if len(mapkeys) == 0 {
			return nil
		}
		mapValue := make(map[string]any)
		for _, key := range mapkeys {
			if key.Kind() != reflect.String {
				return ErrJSUnmarshal{Msg: fmt.Sprintf("unexpected key type %s while parsing map", key.Kind()), Value: value}
			}
			mapValue[key.String()] = rv.MapIndex(key).Interface()
		}
		if err := o.UnmarshalFromGoMap(mapValue); err != nil {
			return oops.Join(ErrJSUnmarshal{Msg: "error parsing map", Value: value}, err)
		}
		j.Object = Some(o)
		return nil

	case reflect.Slice:
		a := JSArray{}
		for i := 0; i < rv.Len(); i++ {
			v := JSValue{}
			err := v.UnmarshalFromGo(rv.Index(i).Interface())
			if err != nil {
				return oops.Join(ErrJSUnmarshal{Msg: "error parsing element of array", Value: value}, err)
			}
			a = append(a, v)
		}
		j.Array = Some(a)
		return nil

	case reflect.Pointer:
		rvElem := rv.Elem()
		if rvElem.IsZero() {
			j.Null = Some(struct{}{})
			return nil
		}
		if err := j.UnmarshalFromGo(rvElem.Interface()); err != nil {
			return oops.Join(ErrJSUnmarshal{Msg: "error parsing pointer", Value: value}, err)
		}
		return nil

	case reflect.String:
		j.String = Some(rv.String())
		return nil

	case reflect.Float32, reflect.Float64, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		converted := rv.Convert(reflect.TypeOf(float64(0)))
		j.Float = Some(converted.Float())
		return nil

	case reflect.Bool:
		j.Boolean = Some(rv.Bool())
		return nil
	}
}

func (j *JSObject) UnmarshalFromGoMap(mapValue map[string]any) error {
	result := make(JSObject)
	for k, v := range mapValue {
		jsv := JSValue{}
		if err := jsv.UnmarshalFromGo(v); err != nil {
			return oops.Join(ErrJSUnmarshal{Msg: "error parsing map", Value: mapValue}, err)
		}
		result[k] = jsv
	}
	*j = result
	return nil
}

// marshal json

type ErrJSMarshal struct {
	Msg   string
	Value JSValue
}

func (e ErrJSMarshal) Error() string {
	return fmt.Sprintf("%s: %v", e.Msg, e.Value)
}

func (j JSValue) MarshalJSON() ([]byte, error) {
	switch {
	case j.Object.IsPresent:
		return json.Marshal(j.Object.Value)
	case j.Array.IsPresent:
		return json.Marshal(j.Array.Value)
	case j.String.IsPresent:
		return json.Marshal(j.String.Value)
	case j.Float.IsPresent:
		return json.Marshal(j.Float.Value)
	case j.Int.IsPresent:
		return json.Marshal(j.Int.Value)
	case j.Boolean.IsPresent:
		return json.Marshal(j.Boolean.Value)
	case j.Null.IsPresent:
		return []byte("null"), nil
	case j.Undefined.IsPresent:
		return []byte("undefined"), nil
	default:
		return nil, ErrJSMarshal{Msg: "error marshaling value", Value: j}
	}
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
	case j.Float.IsPresent:
		return fmt.Sprintf("%f", j.Float.Value)
	case j.Int.IsPresent:
		return fmt.Sprintf("%d", j.Int.Value)
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

func (p JSPath) GoString() string {
	sb := strings.Builder{}
	for _, s := range p {
		sb.WriteString("[")
		sb.WriteString(s.GoString())
		sb.WriteString("]")
	}
	return sb.String()
}
