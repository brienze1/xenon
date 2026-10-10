package sandbox

import (
	"context"
	"strings"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/eventloop"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const nonLegacyMarker = "use sandbox2"

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
	sdk              *goja.Object
	pmInternal       *goja.Object
	legacy           *goja.Object
	loaded           map[string]*goja.Object
	timers           map[int64]func()
	unsupported      map[*goja.Object]bool
	noticed          map[string]bool
	failure          *entity.ScriptError
	script           entity.ScriptRef
	logs             []entity.ScriptLog
	notices          []entity.ScriptLog
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
		run = run.Failed(syntaxError(err, script)).WithEffects(entity.ScriptEffects{Request: scope.Request, Scopes: scope.Scopes, Vault: scope.Vault}).Finished(timeutils.Now())
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
		ctx:         ctx,
		loop:        loop,
		modules:     w.modules,
		loaded:      map[string]*goja.Object{},
		timers:      map[int64]func(){},
		unsupported: map[*goja.Object]bool{},
		noticed:     map[string]bool{},
		script:      run.Script,
		scope:       scope,
		requests:    requests,
		logs:        []entity.ScriptLog{},
		notices:     []entity.ScriptLog{},
		nested:      []entity.NestedRequest{},
	}
	var wrapper *goja.Promise
	var watch *guard
	loop.Run(func(vm *goja.Runtime) {
		e.vm = vm
		watch = limits.watch(loopStopper{vm: vm, loop: loop})
		if err := e.prepare(); err != nil {
			e.fail(err)
			return
		}
		value, err := vm.RunProgram(program)
		if err != nil {
			e.fail(err)
			return
		}
		wrapper, _ = value.Export().(*goja.Promise)
		e.whenSettled(value)
	})
	tripped := watch.stop()
	e.finished = true
	loop.Terminate()
	e.vm.ClearInterrupt()
	effects := e.effects()
	if tripped != "" {
		effects.Request = scope.Request
		return run.Failed(limitError(tripped.String(), run.Script)).WithEffects(effects)
	}
	return e.result(run, wrapper, effects)
}

func (e *execution) prepare() (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if exception, isException := recovered.(*goja.Exception); isException {
				err = exception
				return
			}
			panic(recovered)
		}
	}()
	e.vm.SetMaxCallStackSize(maxCallStackSize)
	e.vm.SetPromiseRejectionTracker(e.trackRejection)
	e.vm.GlobalObject().Delete("require")
	errorConstructor, _ := goja.AssertConstructor(e.vm.Get("Error"))
	e.errorConstructor = errorConstructor
	e.installTimers()
	host := e.host()
	e.installSdk(host)
	e.installGlobals(host)
	canonical := e.installConsole(host)
	e.installRequire()
	e.installPm(host, canonical)
	if !strings.Contains(firstLine(e.script.Text), nonLegacyMarker) {
		e.installLegacy(host)
	}
	return nil
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

func (e *execution) set(name string, value any) {
	if err := e.vm.Set(name, value); err != nil {
		panic(err)
	}
}

func (e *execution) fail(err error) {
	if e.skipped && isSkipInterrupt(err) {
		e.loop.StopNoWait()
		return
	}
	if e.failure == nil {
		scriptError := e.runtimeError(err)
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

func (e *execution) result(run entity.ScriptRun, wrapper *goja.Promise, effects entity.ScriptEffects) entity.ScriptRun {
	if scriptError := e.outcome(wrapper); scriptError != nil {
		effects.Request = e.scope.Request
		return run.Failed(*scriptError).WithEffects(effects)
	}
	run = run.WithEffects(effects)
	if e.skipped {
		return run.SkippedRequest()
	}
	return run
}

func (e *execution) outcome(wrapper *goja.Promise) *entity.ScriptError {
	if e.failure != nil {
		return e.failure
	}
	if e.skipped {
		return nil
	}
	if wrapper != nil && wrapper.State() == goja.PromiseStateRejected {
		scriptError := e.thrownError(wrapper.Result())
		return &scriptError
	}
	for _, rejected := range e.rejections {
		if rejected != wrapper {
			scriptError := e.rejectionError(rejected.Result())
			return &scriptError
		}
	}
	return nil
}
