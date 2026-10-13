package sandbox

import (
	"github.com/dop251/goja"
)

func (e *execution) installGlobals(host *goja.Object) {
	installed := e.shim("globals.js", host, e.vm.GlobalObject())
	for _, name := range installed.Keys() {
		e.set(name, installed.Get(name))
	}
	e.shim("intl_install.js", host, e.vm.GlobalObject())
	e.shim("sugar_install.js", host, e.vm.GlobalObject())
}
