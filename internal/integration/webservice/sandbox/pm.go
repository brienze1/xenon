package sandbox

import (
	"encoding/json"

	"github.com/dop251/goja"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
)

func (e *execution) installPm() {
	lists := e.shim("property_list.js")
	urls := e.shim("url.js", lists)
	shims := e.vm.NewObject()
	for name, value := range map[string]goja.Value{
		"Request":  e.shim("pm_request.js", lists, urls).ToObject(e.vm).Get("Request"),
		"Response": e.shim("pm_response.js", lists).ToObject(e.vm).Get("Response"),
	} {
		_ = shims.Set(name, value)
	}
	bridge := e.vm.NewObject()
	request, _ := json.Marshal(dto.NewScriptRequest(e.scope.Request))
	_ = bridge.Set("request", string(request))
	_ = bridge.Set("variables", e.variablesBridge())
	_ = bridge.Set("skipRequest", e.vm.ToValue(e.skipRequest))
	_ = bridge.Set("send", e.vm.ToValue(e.sendRequest))
	installed := e.shim("pm.js", bridge, shims).ToObject(e.vm)
	e.requestSnapshot, _ = goja.AssertFunction(installed.Get("request"))
	e.set("pm", installed.Get("pm"))
}

func (e *execution) sentRequest() entity.Request {
	snapshot, err := e.requestSnapshot(goja.Undefined())
	if err != nil {
		return e.scope.Request
	}
	var request dto.ScriptRequest
	if json.Unmarshal([]byte(snapshot.String()), &request) != nil {
		return e.scope.Request
	}
	return request.ToEntity(e.scope.Request)
}
