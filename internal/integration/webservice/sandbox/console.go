package sandbox

import (
	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
)

type consoleCall struct {
	level     enums.ConsoleLevel
	arguments []goja.Value
}

func (e *execution) installConsole() {
	console := e.vm.NewObject()
	for _, level := range []enums.ConsoleLevel{consolelevel.Log, consolelevel.Info, consolelevel.Warn, consolelevel.Error, consolelevel.Debug} {
		_ = console.Set(level.String(), func(call goja.FunctionCall) goja.Value {
			e.logs = append(e.logs, consoleCall{level: level, arguments: tail(call.Arguments, 0)})
			return goja.Undefined()
		})
	}
	_ = console.Set("clear", func(goja.FunctionCall) goja.Value {
		e.logs = []consoleCall{}
		return goja.Undefined()
	})
	e.set("console", console)
}
