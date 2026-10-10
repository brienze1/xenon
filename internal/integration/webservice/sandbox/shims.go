package sandbox

import (
	"embed"
	"sync"

	"github.com/dop251/goja"
)

const shimsDir = "shims"

//go:embed shims/*.js
var shimFiles embed.FS

var shimPrograms sync.Map

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

func (e *execution) shim(name string, arguments ...goja.Value) goja.Value {
	factory, err := e.vm.RunProgram(shimProgram(name))
	if err != nil {
		panic(err)
	}
	call, _ := goja.AssertFunction(factory)
	exports, err := call(goja.Undefined(), arguments...)
	if err != nil {
		panic(err)
	}
	return exports
}
