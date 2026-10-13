package sandbox

import (
	"embed"
	"sync"

	"github.com/dop251/goja"
)

const shimsDir = "shims"

//go:embed shims/*.js
var shimFiles embed.FS

var (
	shimPrograms sync.Map
	sdkShims     = []string{"sdk_util.js", "property_list.js", "variable_scope.js", "url.js", "pm_request.js", "pm_response.js", "chai_postman.js"}
	sdkClasses   = []string{
		"PropertyBase", "Property", "PropertyList", "Description", "Variable", "VariableList", "VariableScope", "QueryParam", "Url", "Header",
		"HeaderList", "FormParam", "RequestBody", "RequestAuth", "Request", "Cookie", "CookieList", "Response",
	}
)

func shimProgram(name string) *goja.Program {
	if cached, found := shimPrograms.Load(name); found {
		return cached.(*goja.Program)
	}
	source, err := shimFiles.ReadFile(shimsDir + "/" + name)
	if err != nil {
		panic(err)
	}
	program := goja.MustCompile(name, string(source), true)
	shimPrograms.Store(name, program)
	return program
}

func (e *execution) shim(name string, arguments ...goja.Value) *goja.Object {
	factory, err := e.vm.RunProgram(shimProgram(name))
	if err != nil {
		panic(err)
	}
	call, _ := goja.AssertFunction(factory)
	exports, err := call(goja.Undefined(), arguments...)
	if err != nil {
		panic(err)
	}
	return exports.ToObject(e.vm)
}

func (e *execution) installSdk(host *goja.Object) {
	e.sdk = e.vm.NewObject()
	_ = e.sdk.Set("host", host)
	for _, name := range sdkShims {
		exports := e.shim(name, e.sdk)
		for _, key := range exports.Keys() {
			_ = e.sdk.Set(key, exports.Get(key))
		}
	}
}

func (e *execution) postmanCollection() *goja.Object {
	module := e.vm.NewObject()
	for _, name := range sdkClasses {
		_ = module.Set(name, e.sdk.Get(name))
	}
	return module
}
