package persistence

import (
	"context"
	"slices"
	"strings"
	"sync"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/logger"
)

type activeRunsPersistence struct {
	runs  map[string]*entity.ActiveRun
	mutex sync.Mutex
}

func NewActiveRunsPersistence() adapter.ActiveRunsPersistence {
	return &activeRunsPersistence{runs: map[string]*entity.ActiveRun{}}
}

func (a *activeRunsPersistence) Claim(ctx context.Context, run entity.Run, control *entity.RunControl) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": run.Target.LocalID, "itemId": run.Target.ItemID, "runId": run.ID})

	a.mutex.Lock()
	defer a.mutex.Unlock()
	for _, active := range a.runs {
		if active.Run.Target.LocalID == run.Target.LocalID && active.Run.Target.ItemID == run.Target.ItemID {
			return usecase.RejectActiveRun(active.Run.ID)
		}
	}
	a.runs[activeRunKey(run.Target.LocalID, run.ID)] = &entity.ActiveRun{Run: run, Control: control}

	logger.Debug(ctx, "Finished", map[string]any{"localId": run.Target.LocalID, "itemId": run.Target.ItemID, "runId": run.ID})
	return nil
}

func (a *activeRunsPersistence) Change(ctx context.Context, localId, runId string, change func(*entity.Run)) (*entity.ActiveRun, bool) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	a.mutex.Lock()
	defer a.mutex.Unlock()
	active, found := a.runs[activeRunKey(localId, runId)]
	if !found {
		return nil, false
	}
	change(&active.Run)
	changed := *active

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": changed.Run.Status})
	return &changed, true
}

func (a *activeRunsPersistence) Release(ctx context.Context, localId, runId string) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	a.mutex.Lock()
	delete(a.runs, activeRunKey(localId, runId))
	a.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId})
}

func (a *activeRunsPersistence) Find(ctx context.Context, localId, runId string) (*entity.ActiveRun, bool) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	a.mutex.Lock()
	defer a.mutex.Unlock()
	active, found := a.runs[activeRunKey(localId, runId)]
	if !found {
		return nil, false
	}
	copied := *active

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": copied.Run.Status})
	return &copied, true
}

func (a *activeRunsPersistence) ListWorkspace(ctx context.Context, localId string) []entity.ActiveRun {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	runs := a.list(func(active entity.ActiveRun) bool { return active.Run.Target.LocalID == localId })

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runs": len(runs)})
	return runs
}

func (a *activeRunsPersistence) ListActive(ctx context.Context) []entity.ActiveRun {
	logger.Debug(ctx, "Started")

	runs := a.list(func(entity.ActiveRun) bool { return true })

	logger.Debug(ctx, "Finished", map[string]any{"runs": len(runs)})
	return runs
}

func (a *activeRunsPersistence) list(keep func(entity.ActiveRun) bool) []entity.ActiveRun {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	runs := make([]entity.ActiveRun, 0, len(a.runs))
	for _, active := range a.runs {
		if keep(*active) {
			runs = append(runs, *active)
		}
	}
	slices.SortFunc(runs, func(left, right entity.ActiveRun) int {
		return strings.Compare(left.Run.ID, right.Run.ID)
	})
	return runs
}

func activeRunKey(localId, runId string) string {
	return localId + "/" + runId
}
