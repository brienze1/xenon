package sandbox

import (
	"context"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type scriptRunnerWebservice struct {
	modules *moduleRegistry
	limits  func(entity.AppSettings) limits
}

func NewScriptRunnerWebservice() adapter.ScriptRunnerWebservice {
	return &scriptRunnerWebservice{modules: sharedModuleRegistry(), limits: newLimits}
}

type execution struct {
	loopTimers       loopTimers
	moduleRequire    goja.Value
	vm               *goja.Runtime
	errorConstructor goja.Constructor
	loop             *eventloop.EventLoop
	modules          *moduleRegistry
	loaded           map[string]*goja.Object
	timers           map[int64]func()
	failure          *entity.ScriptError
	script           entity.ScriptRef
	logs             []consoleCall
	rejections       []*goja.Promise
	nextTimer        int64
}

type loopStopper struct {
	vm   *goja.Runtime
	loop *eventloop.EventLoop
}

func (l loopStopper) Interrupt(v any) {
	l.vm.Interrupt(v)
}

func (l loopStopper) StopNoWait() {
	l.loop.StopNoWait()
}

func (w *scriptRunnerWebservice) Run(ctx context.Context, script entity.ScriptRef, settings entity.AppSettings) entity.ScriptRun {
	logger.Debug(ctx, "Started", map[string]any{"script": script.Label()})

	run := entity.NewScriptRun(script, timeutils.Now())
	program, err := goja.Compile(scriptFile, wrap(script.Text), false)
	if err != nil {
		run = run.Failed(syntaxError(err, script)).Finished(timeutils.Now())
		logger.Debug(ctx, "Finished", map[string]any{"script": script.Label(), "status": run.Status})
		return run
	}
	if scriptError := w.execute(program, script, w.limits(settings)); scriptError != nil {
		run = run.Failed(*scriptError)
	}
	run = run.Finished(timeutils.Now())

	logger.Debug(ctx, "Finished", map[string]any{"script": script.Label(), "status": run.Status})
	return run
}

func (w *scriptRunnerWebservice) execute(program *goja.Program, script entity.ScriptRef, limits limits) *entity.ScriptError {
	loop := eventloop.NewEventLoop(eventloop.EnableConsole(false))
	e := &execution{loop: loop, modules: w.modules, loaded: map[string]*goja.Object{}, timers: map[int64]func(){}, script: script}
	var wrapper *goja.Promise
	var watch *guard
	loop.Run(func(vm *goja.Runtime) {
		e.vm = vm
		e.prepare()
		watch = limits.watch(loopStopper{vm: vm, loop: loop})
		value, err := vm.RunProgram(program)
		if err != nil {
			e.fail(err)
			return
		}
		wrapper, _ = value.Export().(*goja.Promise)
	})
	tripped := watch.stop()
	loop.Terminate()
	if tripped != "" {
		scriptError := limitError(tripped.String(), script)
		return &scriptError
	}
	return e.outcome(wrapper)
}

func (e *execution) prepare() {
	e.vm.SetMaxCallStackSize(maxCallStackSize)
	e.vm.SetPromiseRejectionTracker(e.trackRejection)
	e.vm.GlobalObject().Delete("require")
	e.installTimers()
	e.installConsole()
	e.installRequire()
}

func (e *execution) set(name string, value any) {
	if err := e.vm.Set(name, value); err != nil {
		panic(err)
	}
}

func (e *execution) fail(err error) {
	if e.failure == nil {
		scriptError := runtimeError(err, e.script)
		e.failure = &scriptError
	}
	e.loop.StopNoWait()
}

func (e *execution) trackRejection(promise *goja.Promise, operation goja.PromiseRejectionOperation) {
	if operation == goja.PromiseRejectionReject {
		e.rejections = append(e.rejections, promise)
		return
	}
	for index, rejected := range e.rejections {
		if rejected == promise {
			e.rejections = append(e.rejections[:index], e.rejections[index+1:]...)
			return
		}
	}
}

func (e *execution) outcome(wrapper *goja.Promise) *entity.ScriptError {
	if e.failure != nil {
		return e.failure
	}
	if wrapper != nil && wrapper.State() == goja.PromiseStateRejected {
		scriptError := thrownError(wrapper.Result(), e.script)
		return &scriptError
	}
	for _, rejected := range e.rejections {
		if rejected != wrapper {
			scriptError := rejectionError(rejected.Result(), e.script)
			return &scriptError
		}
	}
	return nil
}
