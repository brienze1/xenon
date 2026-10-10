package sandbox

import (
	"github.com/dop251/goja"
)

const lodash3Module = "lodash3"

var lazyGlobals = map[string]func(e *execution) goja.Value{
	"Buffer":   func(e *execution) goja.Value { return e.require("buffer").Get("Buffer") },
	"CryptoJS": func(e *execution) goja.Value { return e.require("crypto-js") },
	"tv4":      func(e *execution) goja.Value { return e.require("tv4") },
	"cheerio":  func(e *execution) goja.Value { return e.require("cheerio") },
	"Backbone": func(e *execution) goja.Value { return e.require("backbone") },
	"_":        legacyLodash,
}

var internalModules = map[string]bool{lodash3Module: true, "intl": true}

func (e *execution) installRequire() {
	errorConstructor, _ := goja.AssertConstructor(e.vm.Get("Error"))
	e.errorConstructor = errorConstructor
	e.moduleRequire = e.vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return e.require(e.moduleName(call))
	})
	e.set("require", func(call goja.FunctionCall) goja.Value {
		name := e.moduleName(call)
		if internalModules[name] {
			panic(e.newError("Cannot find module '" + name + "'"))
		}
		return e.require(name)
	})
	for name, load := range lazyGlobals {
		e.defineLazy(name, load)
	}
}

func (e *execution) moduleName(call goja.FunctionCall) string {
	name := call.Argument(0).String()
	if !e.modules.has(name) {
		panic(e.newError("Cannot find module '" + name + "'"))
	}
	return name
}

func (e *execution) require(name string) *goja.Object {
	if exports, found := e.loaded[name]; found {
		return exports
	}
	program, err := e.modules.program(name)
	if err != nil {
		panic(e.newError("Cannot load module '" + name + "': " + err.Error()))
	}
	module := e.vm.NewObject()
	exports := e.vm.NewObject()
	_ = module.Set("exports", exports)
	e.loaded[name] = exports
	factory, err := e.vm.RunProgram(program)
	if err != nil {
		panic(err)
	}
	call, _ := goja.AssertFunction(factory)
	if _, err = call(goja.Undefined(), exports, e.moduleRequire, module, e.vm.GlobalObject()); err != nil {
		panic(err)
	}
	loaded := module.Get("exports").ToObject(e.vm)
	e.loaded[name] = loaded
	return loaded
}

func (e *execution) defineLazy(name string, load func(e *execution) goja.Value) {
	global := e.vm.GlobalObject()
	getter := e.vm.ToValue(func(goja.FunctionCall) goja.Value {
		_ = global.DefineDataProperty(name, goja.Undefined(), goja.FLAG_TRUE, goja.FLAG_TRUE, goja.FLAG_FALSE)
		value := load(e)
		_ = global.DefineDataProperty(name, value, goja.FLAG_TRUE, goja.FLAG_TRUE, goja.FLAG_FALSE)
		return value
	})
	setter := e.vm.ToValue(func(call goja.FunctionCall) goja.Value {
		_ = global.DefineDataProperty(name, call.Argument(0), goja.FLAG_TRUE, goja.FLAG_TRUE, goja.FLAG_FALSE)
		return goja.Undefined()
	})
	_ = global.DefineAccessorProperty(name, getter, setter, goja.FLAG_TRUE, goja.FLAG_FALSE)
}

func legacyLodash(e *execution) goja.Value {
	lodash := e.require(lodash3Module)
	noConflict, _ := goja.AssertFunction(lodash.Get("noConflict"))
	value, err := noConflict(lodash)
	if err != nil {
		panic(err)
	}
	return value
}

func (e *execution) newError(message string) *goja.Object {
	errorObject, _ := e.errorConstructor(nil, e.vm.ToValue(message))
	return errorObject
}
