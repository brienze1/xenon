package sandbox

import (
	"github.com/dop251/goja"
)

type loopTimers struct {
	setTimeout    goja.Callable
	setInterval   goja.Callable
	clearTimeout  goja.Callable
	clearInterval goja.Callable
}

func (e *execution) installTimers() {
	e.loopTimers = loopTimers{
		setTimeout:    e.callable("setTimeout"),
		setInterval:   e.callable("setInterval"),
		clearTimeout:  e.callable("clearTimeout"),
		clearInterval: e.callable("clearInterval"),
	}
	e.set("setTimeout", func(call goja.FunctionCall) goja.Value {
		return e.schedule(call.Argument(0), call.Argument(1), tail(call.Arguments, 2), false)
	})
	e.set("setInterval", func(call goja.FunctionCall) goja.Value {
		return e.schedule(call.Argument(0), call.Argument(1), tail(call.Arguments, 2), true)
	})
	e.set("setImmediate", func(call goja.FunctionCall) goja.Value {
		return e.schedule(call.Argument(0), e.vm.ToValue(0), tail(call.Arguments, 1), false)
	})
	for _, name := range []string{"clearTimeout", "clearInterval", "clearImmediate"} {
		e.set(name, e.clearTimer)
	}
}

func (e *execution) callable(name string) goja.Callable {
	function, _ := goja.AssertFunction(e.vm.Get(name))
	return function
}

func (e *execution) schedule(callbackValue, delay goja.Value, arguments []goja.Value, repeating bool) goja.Value {
	callback, ok := goja.AssertFunction(callbackValue)
	if !ok {
		panic(e.vm.NewTypeError("The \"callback\" argument must be of type function"))
	}
	e.nextTimer++
	id := e.nextTimer
	fire := e.vm.ToValue(func(goja.FunctionCall) goja.Value {
		if !repeating {
			delete(e.timers, id)
		}
		if _, err := callback(goja.Undefined(), arguments...); err != nil {
			e.fail(err)
		}
		return goja.Undefined()
	})
	schedule, clear := e.loopTimers.setTimeout, e.loopTimers.clearTimeout
	if repeating {
		schedule, clear = e.loopTimers.setInterval, e.loopTimers.clearInterval
	}
	handle, err := schedule(goja.Undefined(), fire, delay)
	if err != nil {
		panic(err)
	}
	e.timers[id] = func() { _, _ = clear(goja.Undefined(), handle) }
	return e.vm.ToValue(id)
}

func (e *execution) clearTimer(call goja.FunctionCall) goja.Value {
	id := call.Argument(0).ToInteger()
	if cancel, found := e.timers[id]; found {
		delete(e.timers, id)
		cancel()
	}
	return goja.Undefined()
}

func tail(values []goja.Value, from int) []goja.Value {
	return append([]goja.Value{}, values[min(from, len(values)):]...)
}
