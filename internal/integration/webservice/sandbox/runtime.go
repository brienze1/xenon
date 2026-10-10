package sandbox

import (
	"context"
	"maps"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"

	"github.com/brienze1/xenon/internal/application/usecase"
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
	ctx              context.Context
	requests         usecase.SendScriptRequest
	moduleRequire    goja.Value
	keepAlive        goja.Value
	vm               *goja.Runtime
	errorConstructor goja.Constructor
	loop             *eventloop.EventLoop
	modules          *moduleRegistry
	loaded           map[string]*goja.Object
	timers           map[int64]func()
	failure          *entity.ScriptError
	local            map[string]any
	requestSnapshot  goja.Callable
	jsonStringify    goja.Callable
	script           entity.ScriptRef
	logs             []entity.ScriptLog
	nested           []entity.NestedRequest
	rejections       []*goja.Promise
	scope            entity.ScriptContext
	nextTimer        int64
	pendingRequests  int
	skipped          bool
	finished         bool
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

func (w *scriptRunnerWebservice) Run(ctx context.Context, script entity.ScriptRef, scope entity.ScriptContext, requests usecase.SendScriptRequest, settings entity.AppSettings) entity.ScriptRun {
	logger.Debug(ctx, "Started", map[string]any{"script": script.Label()})

	run := entity.NewScriptRun(script, timeutils.Now())
	program, err := goja.Compile(scriptFile, wrap(script.Text), false)
	if err != nil {
		run = run.Failed(syntaxError(err, script)).WithEffects(scope.Request, maps.Clone(scope.LocalVariables), nil, nil).Finished(timeutils.Now())
		logger.Debug(ctx, "Finished", map[string]any{"script": script.Label(), "status": run.Status})
		return run
	}
	run = w.execute(ctx, program, run, scope, requests, w.limits(settings)).Finished(timeutils.Now())

	logger.Debug(ctx, "Finished", map[string]any{"script": script.Label(), "status": run.Status})
	return run
}

func (w *scriptRunnerWebservice) execute(ctx context.Context, program *goja.Program, run entity.ScriptRun, scope entity.ScriptContext, requests usecase.SendScriptRequest, limits limits) entity.ScriptRun {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	loop := eventloop.NewEventLoop(eventloop.EnableConsole(false))
	e := &execution{
		ctx:      ctx,
		loop:     loop,
		modules:  w.modules,
		loaded:   map[string]*goja.Object{},
		timers:   map[int64]func(){},
		script:   run.Script,
		scope:    scope,
		local:    maps.Clone(scope.LocalVariables),
		requests: requests,
		logs:     []entity.ScriptLog{},
		nested:   []entity.NestedRequest{},
	}
	if e.local == nil {
		e.local = map[string]any{}
	}
	var wrapper *goja.Promise
	var watch *guard
	loop.Run(func(vm *goja.Runtime) {
		e.vm = vm
		watch = limits.watch(loopStopper{vm: vm, loop: loop})
		e.prepare()
		value, err := vm.RunProgram(program)
		if err != nil {
			e.fail(err)
			return
		}
		wrapper, _ = value.Export().(*goja.Promise)
	})
	tripped := watch.stop()
	e.finished = true
	loop.Terminate()
	if tripped != "" {
		return run.Failed(limitError(tripped.String(), run.Script)).WithEffects(scope.Request, e.local, e.logs, e.nested)
	}
	return e.result(run, wrapper)
}

func (e *execution) prepare() {
	e.vm.SetMaxCallStackSize(maxCallStackSize)
	e.vm.SetPromiseRejectionTracker(e.trackRejection)
	e.vm.GlobalObject().Delete("require")
	e.jsonStringify, _ = goja.AssertFunction(e.vm.Get("JSON").ToObject(e.vm).Get("stringify"))
	e.installTimers()
	e.installConsole()
	e.installRequire()
	e.installPm()
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

func (e *execution) result(run entity.ScriptRun, wrapper *goja.Promise) entity.ScriptRun {
	if scriptError := e.outcome(wrapper); scriptError != nil {
		return run.Failed(*scriptError).WithEffects(e.scope.Request, e.local, e.logs, e.nested)
	}
	run = run.WithEffects(e.sentRequest(), e.local, e.logs, e.nested)
	if e.skipped {
		return run.SkippedRequest()
	}
	return run
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
