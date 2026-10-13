package sandbox

import (
	"encoding/json"

	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
)

const keepAliveMs = 3_600_000

func (e *execution) sendRequest(call goja.FunctionCall) goja.Value {
	if e.requests == nil {
		panic(e.newError("pm.sendRequest is not available in this run"))
	}
	var definition dto.ScriptRequest
	if err := json.Unmarshal([]byte(call.Argument(0).String()), &definition); err != nil {
		panic(e.vm.NewTypeError("pm.sendRequest: the request is not valid"))
	}
	scopes := e.currentScopes(call.Argument(1))
	done, _ := goja.AssertFunction(call.Argument(2))
	request := definition.ToEntity(entity.NewRequest("", ""))
	e.holdLoop()
	go func() {
		nested := e.requests.Send(e.ctx, request, scopes)
		e.loop.RunOnLoop(func(*goja.Runtime) {
			e.completeRequest(nested, done)
		})
	}()
	return goja.Undefined()
}

func (e *execution) completeRequest(nested entity.NestedRequest, done goja.Callable) {
	if e.finished {
		return
	}
	e.nested = append(e.nested, nested)
	e.releaseLoop()
	var err error
	if nested.HasFailed() {
		_, err = done(goja.Undefined(), e.vm.ToValue(nested.FailureMessage()), goja.Null())
	} else {
		data, _ := json.Marshal(dto.NewScriptResponse(*nested.Response))
		_, err = done(goja.Undefined(), goja.Null(), e.vm.ToValue(string(data)))
	}
	if err != nil {
		e.fail(err)
	}
}

func (e *execution) holdLoop() {
	if e.pendingRequests == 0 {
		noop := e.vm.ToValue(func(goja.FunctionCall) goja.Value { return goja.Undefined() })
		e.keepAlive, _ = e.loopTimers.setInterval(goja.Undefined(), noop, e.vm.ToValue(keepAliveMs))
	}
	e.pendingRequests++
}

func (e *execution) releaseLoop() {
	e.pendingRequests--
	if e.pendingRequests == 0 {
		_, _ = e.loopTimers.clearInterval(goja.Undefined(), e.keepAlive)
	}
}
