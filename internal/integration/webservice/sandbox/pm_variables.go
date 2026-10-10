package sandbox

import (
	"encoding/json"

	"github.com/dop251/goja"
)

func (e *execution) variablesBridge() *goja.Object {
	variables := e.vm.NewObject()
	_ = variables.Set("has", func(call goja.FunctionCall) goja.Value {
		_, found := e.variable(call.Argument(0).String())
		return e.vm.ToValue(found)
	})
	_ = variables.Set("get", func(call goja.FunctionCall) goja.Value {
		value, found := e.variable(call.Argument(0).String())
		if !found {
			return goja.Undefined()
		}
		encoded, _ := json.Marshal(value)
		return e.vm.ToValue(string(encoded))
	})
	_ = variables.Set("set", func(call goja.FunctionCall) goja.Value {
		var value any
		_ = json.Unmarshal([]byte(call.Argument(1).String()), &value)
		e.local[call.Argument(0).String()] = value
		return goja.Undefined()
	})
	_ = variables.Set("unset", func(call goja.FunctionCall) goja.Value {
		delete(e.local, call.Argument(0).String())
		return goja.Undefined()
	})
	_ = variables.Set("clear", func(goja.FunctionCall) goja.Value {
		clear(e.local)
		return goja.Undefined()
	})
	_ = variables.Set("replaceIn", func(call goja.FunctionCall) goja.Value {
		return e.vm.ToValue(e.scope.Variables.WithLocal(e.local).Resolve(call.Argument(0).String(), nil))
	})
	_ = variables.Set("toObject", func(goja.FunctionCall) goja.Value {
		values := map[string]any{}
		for _, variable := range e.scope.Variables.WithLocal(e.local).Variables() {
			if variable.IsResolvable() {
				values[variable.Name] = variable.Value
			}
		}
		for name, value := range e.local {
			values[name] = value
		}
		encoded, _ := json.Marshal(values)
		return e.vm.ToValue(string(encoded))
	})
	return variables
}

func (e *execution) variable(name string) (any, bool) {
	if value, found := e.local[name]; found {
		return value, true
	}
	variable, found := e.scope.Variables.Lookup(name)
	if !found || !variable.IsResolvable() {
		return nil, false
	}
	return variable.Value, true
}
