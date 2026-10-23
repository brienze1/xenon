package publisher

import (
	"context"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

const runProgressInterval = 100 * time.Millisecond

type throttledRun struct {
	emittedAt time.Time
	pending   *dto.RunProgressChanged
	timer     *time.Timer
}

type runEventPublisher struct {
	eventEmitter adapter.EventEmitter
	runs         map[string]*throttledRun
	mutex        sync.Mutex
}

func NewRunEvent(eventEmitter adapter.EventEmitter) adapter.RunEventPublisher {
	return &runEventPublisher{eventEmitter: eventEmitter, runs: map[string]*throttledRun{}}
}

func (r *runEventPublisher) Progress(ctx context.Context, run entity.Run) {
	logger.Debug(ctx, "Started", map[string]any{"runId": run.ID, "status": run.Status, "completed": run.Progress.Completed})

	event := dto.NewRunProgressChanged(run)
	r.mutex.Lock()
	throttled, found := r.runs[run.ID]
	if !found {
		throttled = &throttledRun{}
		r.runs[run.ID] = throttled
	}
	wait := runProgressInterval - time.Since(throttled.emittedAt)
	switch {
	case throttled.timer != nil:
		throttled.pending = &event
	case wait <= 0:
		r.emitLocked(throttled, event)
	default:
		throttled.pending = &event
		throttled.timer = time.AfterFunc(wait, func() { r.flush(run.ID) })
	}
	r.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"runId": run.ID})
}

func (r *runEventPublisher) Finished(ctx context.Context, run entity.Run) {
	logger.Debug(ctx, "Started", map[string]any{"runId": run.ID, "status": run.Status})

	r.mutex.Lock()
	if throttled, found := r.runs[run.ID]; found {
		if throttled.timer != nil {
			throttled.timer.Stop()
		}
		if throttled.pending != nil {
			r.eventEmitter.Emit(dto.RunProgressChangedEvent, *throttled.pending)
		}
		delete(r.runs, run.ID)
	}
	r.eventEmitter.Emit(dto.RunFinishedEvent, dto.NewRunFinished(run))
	r.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"runId": run.ID})
}

func (r *runEventPublisher) flush(runId string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	throttled, found := r.runs[runId]
	if !found || throttled.pending == nil {
		return
	}
	r.emitLocked(throttled, *throttled.pending)
}

func (r *runEventPublisher) emitLocked(throttled *throttledRun, event dto.RunProgressChanged) {
	r.eventEmitter.Emit(dto.RunProgressChangedEvent, event)
	throttled.emittedAt = time.Now()
	throttled.pending = nil
	throttled.timer = nil
}
