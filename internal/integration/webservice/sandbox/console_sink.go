package sandbox

import (
	"encoding/json"

	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
)

func (e *execution) installConsole(host *goja.Object) goja.Value {
	installed := e.shim("console.js", host)
	e.set("console", installed.Get("console"))
	return installed.Get("canonical")
}

func (e *execution) log(call goja.FunctionCall) goja.Value {
	if e.finished {
		return goja.Undefined()
	}
	var values []any
	_ = json.Unmarshal([]byte(call.Argument(1).String()), &values)
	texts := make([]string, 0, len(values))
	_ = e.vm.ExportTo(call.Argument(2), &texts)
	for index := range texts {
		texts[index] = v8Message(texts[index], e.script.Text)
	}
	log := entity.NewScriptLog(consolelevel.GetConsoleLevel(call.Argument(0).String()), texts, normalizedValues(values, e.script.Text))
	e.logs = append(e.logs, log)
	e.forward(e.withScopes(log))
	return goja.Undefined()
}

func (e *execution) notice(call goja.FunctionCall) goja.Value {
	text := call.Argument(1).String()
	if e.noticed[text] {
		return goja.Undefined()
	}
	e.noticed[text] = true
	notice := entity.NewScriptLog(consolelevel.GetConsoleLevel(call.Argument(0).String()), []string{text}, []any{text})
	e.notices = append(e.notices, notice)
	e.forward(notice)
	return goja.Undefined()
}

func (e *execution) withScopes(log entity.ScriptLog) entity.ScriptLog {
	if e.pmInternal == nil {
		return log
	}
	current, isFunction := goja.AssertFunction(e.pmInternal.Get("scopes"))
	if !isFunction {
		return log
	}
	scopes, err := current(e.pmInternal)
	if err != nil {
		return log
	}
	return log.WithScopes(e.currentScopes(scopes))
}

func (e *execution) forward(log entity.ScriptLog) {
	if e.scope.Console != nil {
		e.scope.Console(log)
	}
}

func normalizedValues(values []any, script string) []any {
	for index, value := range values {
		values[index] = normalizedValue(value, script)
	}
	return values
}

func normalizedValue(value any, script string) any {
	switch typed := value.(type) {
	case string:
		return v8Message(typed, script)
	case []any:
		return normalizedValues(typed, script)
	case map[string]any:
		for key, item := range typed {
			typed[key] = normalizedValue(item, script)
		}
		return typed
	default:
		return value
	}
}
