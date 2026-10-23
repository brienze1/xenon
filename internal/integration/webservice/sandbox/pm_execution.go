package sandbox

import (
	"encoding/json"
	"errors"

	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
)

const skipSignal = "pm.execution.skipRequest"

type runRequestOptions struct {
	Variables map[string]any `json:"variables"`
}

func (e *execution) skipRequest(goja.FunctionCall) goja.Value {
	e.skipped = true
	e.vm.Interrupt(skipSignal)
	e.loop.StopNoWait()
	return goja.Undefined()
}

func (e *execution) setNextRequest(call goja.FunctionCall) goja.Value {
	target := call.Argument(0)
	next := entity.NewStopNextRequest()
	if !goja.IsNull(target) && !goja.IsUndefined(target) {
		next = entity.NewNextRequest(target.String())
	}
	e.nextRequest = &next
	return goja.Undefined()
}

func isSkipInterrupt(err error) bool {
	var interrupted *goja.InterruptedError
	return errors.As(err, &interrupted) && interrupted.Value() == skipSignal
}

func (e *execution) runRequest(call goja.FunctionCall) goja.Value {
	if e.requests == nil {
		panic(e.newError("pm.execution.runRequest is not available in this run"))
	}
	requestId := call.Argument(0).String()
	var options runRequestOptions
	_ = json.Unmarshal([]byte(call.Argument(1).String()), &options)
	scopes := e.currentScopes(call.Argument(2))
	for name, value := range options.Variables {
		scopes.Local = scopes.Local.WithValue(name, value)
	}
	done, _ := goja.AssertFunction(call.Argument(3))
	e.holdLoop()
	go func() {
		nested := e.requests.Run(e.ctx, requestId, scopes)
		e.loop.RunOnLoop(func(*goja.Runtime) {
			e.completeRun(nested, done)
		})
	}()
	return goja.Undefined()
}

func (e *execution) completeRun(nested entity.NestedRequest, done goja.Callable) {
	if e.finished {
		return
	}
	e.nested = append(e.nested, nested)
	e.releaseLoop()
	var err error
	switch {
	case nested.HasFailed():
		_, err = done(goja.Undefined(), e.vm.ToValue(nested.FailureMessage()), goja.Null(), goja.Null())
	case nested.Skipped || nested.Response == nil:
		_, err = done(goja.Undefined(), goja.Null(), goja.Null(), e.vm.ToValue(scopesJSON(nested.Scopes)))
	default:
		response, _ := json.Marshal(dto.NewScriptResponse(*nested.Response))
		_, err = done(goja.Undefined(), goja.Null(), e.vm.ToValue(string(response)), e.vm.ToValue(scopesJSON(nested.Scopes)))
	}
	if err != nil {
		e.fail(err)
	}
}

func scopesJSON(scopes entity.ScriptScopes) string {
	encoded, _ := json.Marshal(dto.NewScriptSandboxScopes(scopes))
	return string(encoded)
}
