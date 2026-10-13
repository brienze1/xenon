package sandbox

import (
	"github.com/dop251/goja"
)

var legacyLibraries = map[string]func(e *execution) goja.Value{
	"_":        legacyLodash,
	"CryptoJS": func(e *execution) goja.Value { return e.require("crypto-js") },
	"tv4":      func(e *execution) goja.Value { return e.require("tv4") },
	"cheerio":  func(e *execution) goja.Value { return e.require("cheerio") },
	"Backbone": func(e *execution) goja.Value { return e.require("backbone") },
}

func (e *execution) installLegacy(host *goja.Object) {
	lazy := e.vm.NewObject()
	for name, load := range legacyLibraries {
		_ = lazy.Set(name, func(goja.FunctionCall) goja.Value { return load(e) })
	}
	e.legacy = e.shim("legacy.js", e.sdk, host, e.vm.Get("pm"), e.vm.Get("console"), e.vm.GlobalObject(), lazy, e.pmInternal.Get("record"))
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
