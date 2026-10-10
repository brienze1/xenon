package sandbox

import "github.com/dop251/goja"

func (e *execution) skipRequest(goja.FunctionCall) goja.Value {
	if e.script.IsPreRequest() {
		e.skipped = true
	}
	return goja.Undefined()
}
