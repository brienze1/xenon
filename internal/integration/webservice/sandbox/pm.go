package sandbox

import (
	"encoding/json"

	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
)

func (e *execution) installPm(host *goja.Object, canonical goja.Value) {
	encoded, _ := json.Marshal(dto.NewScriptSandboxContext(e.script, e.scope))
	e.pmInternal = e.shim("pm.js", e.sdk, host, e.parseJSON(string(encoded)), canonical)
	e.set("pm", e.pmInternal.Get("pm"))
}

func (e *execution) parseJSON(text string) goja.Value {
	parse, _ := goja.AssertFunction(e.vm.Get("JSON").ToObject(e.vm).Get("parse"))
	parsed, err := parse(goja.Undefined(), e.vm.ToValue(text))
	if err != nil {
		panic(err)
	}
	return parsed
}

func (e *execution) whenSettled(wrapper goja.Value) {
	promise, isObject := wrapper.(*goja.Object)
	if !isObject || e.legacy == nil {
		return
	}
	then, _ := goja.AssertFunction(promise.Get("then"))
	finish := e.vm.ToValue(func(goja.FunctionCall) goja.Value {
		e.finishLegacy()
		return goja.Undefined()
	})
	if _, err := then(promise, finish, finish); err != nil {
		e.fail(err)
	}
}

func (e *execution) finishLegacy() {
	finish, _ := goja.AssertFunction(e.legacy.Get("finish"))
	if _, err := finish(e.legacy); err != nil {
		e.fail(err)
	}
}

func (e *execution) effects() entity.ScriptEffects {
	effects := entity.ScriptEffects{Request: e.scope.Request, Scopes: e.scope.Scopes, Vault: e.scope.Vault, Logs: e.logs, Notices: e.notices, Nested: e.nested, Tests: []entity.TestResult{}, NextRequest: e.nextRequest}
	if e.pmInternal == nil {
		return effects
	}
	resultOf, _ := goja.AssertFunction(e.pmInternal.Get("result"))
	encoded, err := resultOf(e.pmInternal)
	if err != nil {
		return effects
	}
	var result dto.ScriptSandboxResult
	if json.Unmarshal([]byte(encoded.String()), &result) != nil {
		return effects
	}
	if result.Request != nil {
		effects.Request = result.Request.ToEntity(e.scope.Request)
	}
	effects.Scopes = result.Scopes.ToEntity(e.scope.Scopes)
	effects.Vault = result.VaultEntity()
	effects.Tests = e.normalizedTests(result.TestsEntity())
	return effects
}

func (e *execution) normalizedTests(tests []entity.TestResult) []entity.TestResult {
	for index := range tests {
		tests[index].ErrorMessage = v8Message(tests[index].ErrorMessage, e.script.Text)
	}
	return tests
}

func (e *execution) currentScopes(value goja.Value) entity.ScriptScopes {
	var scopes dto.ScriptSandboxScopes
	if json.Unmarshal([]byte(value.String()), &scopes) != nil {
		return e.scope.Scopes
	}
	return scopes.ToEntity(e.scope.Scopes)
}
