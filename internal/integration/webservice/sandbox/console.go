package sandbox

import (
	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
)

func (e *execution) installConsole() {
	console := e.vm.NewObject()
	for _, level := range []enums.ConsoleLevel{consolelevel.Log, consolelevel.Info, consolelevel.Warn, consolelevel.Error, consolelevel.Debug} {
		_ = console.Set(level.String(), func(call goja.FunctionCall) goja.Value {
			arguments := make([]string, 0, len(call.Arguments))
			for _, argument := range call.Arguments {
				arguments = append(arguments, e.describe(argument))
			}
			e.logs = append(e.logs, entity.ScriptLog{Level: level, Arguments: arguments})
			return goja.Undefined()
		})
	}
	_ = console.Set("clear", func(goja.FunctionCall) goja.Value {
		e.logs = []entity.ScriptLog{}
		return goja.Undefined()
	})
	e.set("console", console)
}

func (e *execution) describe(value goja.Value) string {
	object, isObject := value.(*goja.Object)
	if !isObject || object.ClassName() == "Error" || object.ClassName() == "Function" {
		return valueText(value)
	}
	encoded, err := e.jsonStringify(goja.Undefined(), object)
	if err != nil || goja.IsUndefined(encoded) {
		return valueText(value)
	}
	return encoded.String()
}
