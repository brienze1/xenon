package sandbox

import (
	"github.com/dop251/goja"
)

func (e *execution) host() *goja.Object {
	host := e.vm.NewObject()
	functions := map[string]func(goja.FunctionCall) goja.Value{
		"load":          e.load,
		"dynamic":       e.dynamic,
		"unsupported":   e.unsupportedApi,
		"log":           e.log,
		"notice":        e.notice,
		"skip":          e.skipRequest,
		"send":          e.sendRequest,
		"run":           e.runRequest,
		"uuid":          e.randomUUID,
		"randomBytes":   e.randomBytes,
		"digest":        e.digest,
		"hmac":          e.hmac,
		"hmacEqual":     e.hmacEqual,
		"atob":          e.atob,
		"btoa":          e.btoa,
		"encodeUtf8":    e.encodeUtf8,
		"decodeUtf8":    e.decodeUtf8,
		"decodeBytes":   e.decodeBytes,
		"decodeBinary":  e.decodeBinary,
		"byteLength":    e.byteLength,
		"base64":        e.base64,
		"domainToASCII": e.domainToASCII,
		"timeZone":      e.timeZone,
	}
	for name, function := range functions {
		_ = host.Set(name, function)
	}
	return host
}

func (e *execution) load(call goja.FunctionCall) goja.Value {
	return e.require(call.Argument(0).String())
}

func (e *execution) dynamic(call goja.FunctionCall) goja.Value {
	if e.scope.Dynamic == nil {
		return goja.Undefined()
	}
	value, found := e.scope.Dynamic(call.Argument(0).String())
	if !found {
		return goja.Undefined()
	}
	return e.vm.ToValue(value)
}

func (e *execution) unsupportedApi(call goja.FunctionCall) goja.Value {
	thrown := e.newError(call.Argument(0).String() + " is not supported in Xenon")
	e.unsupported[thrown] = true
	return thrown
}
