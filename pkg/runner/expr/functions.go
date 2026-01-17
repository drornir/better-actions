package expr

import (
	"strings"

	"github.com/samber/oops"
)

type (
	Function      func(args ...JSValue) (JSValue, error)
	FunctionStore map[string]Function
)

var DefaultFunctions = FunctionStore{}

func init() {
	DefaultFunctions.Add("contains", funcTODOUnimplemented)
	DefaultFunctions.Add("startsWith", funcTODOUnimplemented)
	DefaultFunctions.Add("endsWith", funcTODOUnimplemented)
	DefaultFunctions.Add("format", funcTODOUnimplemented)
	DefaultFunctions.Add("join", funcTODOUnimplemented)
	DefaultFunctions.Add("toJSON", funcTODOUnimplemented)
	DefaultFunctions.Add("fromJSON", funcTODOUnimplemented)
	DefaultFunctions.Add("hashFiles", funcTODOUnimplemented)
	DefaultFunctions.Add("success", funcTODOUnimplemented)
	DefaultFunctions.Add("always", funcTODOUnimplemented)
	DefaultFunctions.Add("cancelled", funcTODOUnimplemented)
	DefaultFunctions.Add("failure", funcTODOUnimplemented)
}

func (fs FunctionStore) Add(name string, f Function) {
	fs[strings.ToLower(name)] = f
}

func (fs FunctionStore) Get(name string) (Function, bool) {
	f, ok := fs[strings.ToLower(name)]
	return f, ok
}

func funcTODOUnimplemented(args ...JSValue) (JSValue, error) {
	return JSValue{}, oops.Errorf("TODO: function is not implemented")
}
