package entity

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
)

type Run struct {
	StartedAt       time.Time
	FinishedAt      time.Time
	Status          enums.RunStatus
	ID              string
	EnvironmentName string
	DataFileName    string
	Target          RunTarget
	Config          RunConfig
	Rows            []RunRowResult
	Notices         []string
	Progress        RunProgress
	Summary         RunSummary
}

func NewRun(id string, target RunTarget, config RunConfig, planned int, startedAt time.Time) Run {
	return Run{
		ID:        id,
		Target:    target,
		Config:    config,
		Status:    runstatus.Running,
		StartedAt: startedAt,
		Progress:  NewRunProgress(planned),
		Rows:      []RunRowResult{},
		Notices:   []string{},
	}
}

func (r Run) IsActive() bool {
	return r.Status == runstatus.Running || r.Status == runstatus.Paused
}

func (r Run) IsPaused() bool {
	return r.Status == runstatus.Paused
}

func (r Run) IsFinished() bool {
	return !r.IsActive()
}

func (r Run) CanPause() bool {
	return r.Status == runstatus.Running && !r.Config.IsPerformance()
}

func (r Run) CanResume() bool {
	return r.Status == runstatus.Paused
}

func (r *Run) Pause() bool {
	if !r.CanPause() {
		return false
	}
	r.Status = runstatus.Paused
	return true
}

func (r *Run) Resume() bool {
	if !r.CanResume() {
		return false
	}
	r.Status = runstatus.Running
	return true
}

func (r *Run) Starting(row RunPlanRow, iteration int, at time.Time) {
	r.Progress = r.Progress.Starting(row, iteration)
	r.Progress.ElapsedMs = r.elapsed(at)
}

func (r *Run) Record(row RunRowResult, at time.Time) {
	r.Rows = append(r.Rows, row)
	r.Progress = r.Progress.After(row)
	r.Progress.ElapsedMs = r.elapsed(at)
}

// Finish is a no-op on a run that already ended.
func (r *Run) Finish(status enums.RunStatus, at time.Time) {
	if r.IsFinished() {
		return
	}
	r.Status = status
	r.FinishedAt = at
	r.Progress.ElapsedMs = r.elapsed(at)
	r.Summary = NewRunSummary(r.Iterations(), r.Rows, r.Progress.ElapsedMs)
}

func (r Run) Iterations() int {
	if r.Config.IsPerformance() {
		return 0
	}
	return r.Config.Functional.Iterations
}

func (r Run) elapsed(at time.Time) int64 {
	return at.Sub(r.StartedAt).Milliseconds()
}

// ResponseKeys are the keys of the responses the run kept for this app session.
func (r Run) ResponseKeys() []string {
	keys := make([]string, 0)
	for _, row := range r.Rows {
		if row.ResponseKey != "" {
			keys = append(keys, row.ResponseKey)
		}
	}
	return keys
}

func (r Run) Row(index int) (RunRowResult, bool) {
	for _, row := range r.Rows {
		if row.Index == index {
			return row, true
		}
	}
	return RunRowResult{}, false
}
