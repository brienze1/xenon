package persistence

import (
	"context"
	"slices"
	"strings"
	"sync"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/logger"
)

type activeRunsPersistence struct {
	runs  map[string]*usecase.ActiveRun
	mutex sync.Mutex
}

func NewActiveRunsPersistence() adapter.ActiveRunsPersistence {
	return &activeRunsPersistence{runs: map[string]*usecase.ActiveRun{}}
}

func (a *activeRunsPersistence) Claim(ctx context.Context, run entity.Run, control usecase.RunControl) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": run.Target.LocalID, "itemId": run.Target.ItemID, "runId": run.ID})

	a.mutex.Lock()
	defer a.mutex.Unlock()
	for _, active := range a.runs {
		if active.Run.Target.Equals(run.Target) {
			return active.Run.ActiveConflict()
		}
	}
	a.runs[activeRunKey(run.Target.LocalID, run.ID)] = &usecase.ActiveRun{Run: run, Control: control}

	logger.Debug(ctx, "Finished", map[string]any{"localId": run.Target.LocalID, "itemId": run.Target.ItemID, "runId": run.ID})
	return nil
}

func (a *activeRunsPersistence) Change(ctx context.Context, localId, runId string, change func(*entity.Run)) (*entity.Run, bool) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	a.mutex.Lock()
	defer a.mutex.Unlock()
	active, found := a.runs[activeRunKey(localId, runId)]
	if !found {
		return nil, false
	}
	change(&active.Run)
	changed := active.Run

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": changed.Status})
	return &changed, true
}

func (a *activeRunsPersistence) Release(ctx context.Context, localId, runId string) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	a.mutex.Lock()
	delete(a.runs, activeRunKey(localId, runId))
	a.mutex.Unlock()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId})
}

func (a *activeRunsPersistence) FindById(ctx context.Context, localId, runId string) (*entity.Run, bool) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	active, found := a.find(localId, runId)
	if !found {
		return nil, false
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": active.Run.Status})
	return &active.Run, true
}

func (a *activeRunsPersistence) FindByWorkspace(ctx context.Context, localId string) []entity.Run {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	active := a.list(func(active usecase.ActiveRun) bool { return active.Run.Target.InWorkspace(localId) })
	runs := make([]entity.Run, 0, len(active))
	for _, run := range active {
		runs = append(runs, run.Run)
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runs": len(runs)})
	return runs
}

func (a *activeRunsPersistence) FindControlById(ctx context.Context, localId, runId string) (*usecase.ActiveRun, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	active, found := a.find(localId, runId)
	if !found {
		return nil, errs.NotFoundError("no active or stored run with the id "+runId, "RUN-01404")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": active.Run.Status})
	return active, nil
}

func (a *activeRunsPersistence) FindControlsByWorkspace(ctx context.Context, localId string) []usecase.ActiveRun {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	runs := a.list(func(active usecase.ActiveRun) bool { return active.Run.Target.InWorkspace(localId) })

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "runs": len(runs)})
	return runs
}

func (a *activeRunsPersistence) FindAllControls(ctx context.Context) []usecase.ActiveRun {
	logger.Debug(ctx, "Started")

	runs := a.list(func(usecase.ActiveRun) bool { return true })

	logger.Debug(ctx, "Finished", map[string]any{"runs": len(runs)})
	return runs
}

func (a *activeRunsPersistence) find(localId, runId string) (*usecase.ActiveRun, bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	active, found := a.runs[activeRunKey(localId, runId)]
	if !found {
		return nil, false
	}
	copied := *active
	return &copied, true
}

func (a *activeRunsPersistence) list(keep func(usecase.ActiveRun) bool) []usecase.ActiveRun {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	runs := make([]usecase.ActiveRun, 0, len(a.runs))
	for _, active := range a.runs {
		if keep(*active) {
			runs = append(runs, *active)
		}
	}
	slices.SortFunc(runs, func(left, right usecase.ActiveRun) int {
		return strings.Compare(left.Run.ID, right.Run.ID)
	})
	return runs
}

func activeRunKey(localId, runId string) string {
	return localId + "/" + runId
}
