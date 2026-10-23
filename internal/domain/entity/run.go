package entity

import (
	"slices"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runexportformat"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
	"github.com/brienze1/xenon/pkg/errs"
)

const PerformanceNextRequestNotice = "pm.execution.setNextRequest is ignored in performance runs"

const UnsavedRunCookiesNotice = "The cookies of the run could not be saved to the workspace"

const StoredRunsLimit = 25

// RunsBeyondStoredLimit are the stored runs past the per-target limit, given newest first: the oldest ones, which are pruned.
func RunsBeyondStoredLimit(runs []Run) []Run {
	if len(runs) <= StoredRunsLimit {
		return []Run{}
	}
	return append([]Run{}, runs[StoredRunsLimit:]...)
}

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
	Performance     *PerformanceReport
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

func (r *Run) RecordSource(environmentName, dataFileName string) {
	r.EnvironmentName = environmentName
	r.DataFileName = dataFileName
}

func (r *Run) IsActive() bool {
	return r.Status == runstatus.Running || r.Status == runstatus.Paused
}

func (r *Run) IsPaused() bool {
	return r.Status == runstatus.Paused
}

func (r *Run) IsFinished() bool {
	return !r.IsActive()
}

func (r *Run) IsPerformance() bool {
	return r.Config.IsPerformance()
}

func (r *Run) ValidateExportAs(format enums.RunExportFormat) error {
	if format == runexportformat.Junit && r.IsPerformance() {
		return invalidRunSetting("format", "JUnit is not available for performance runs", "RUN-04400")
	}
	return nil
}

func (r *Run) ValidateSteering() error {
	if r.IsPerformance() {
		return invalidRunSetting("type", "a performance run cannot pause", "RUN-04400")
	}
	return nil
}

func (r *Run) ActiveConflict() error {
	return errs.ConflictError("a run is already active on this collection or folder", "RUN-02409",
		errs.ErrorDetails{Attribute: "run_id", Messages: []string{r.ID}})
}

func (r *Run) CanPause() bool {
	return r.Status == runstatus.Running && !r.Config.IsPerformance()
}

func (r *Run) CanResume() bool {
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

func (r *Run) Measure(report PerformanceReport, metrics PerformanceMetrics, at time.Time) {
	r.Performance = &report
	r.Progress = r.Progress.Measured(metrics)
	r.Progress.ElapsedMs = r.elapsed(at)
}

// Notice adds a run-level notice once.
func (r *Run) Notice(notice string) {
	if !slices.Contains(r.Notices, notice) {
		r.Notices = append(r.Notices, notice)
	}
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
	if r.Performance != nil {
		r.Summary = r.Summary.WithPerformance(r.Performance.Overall)
	}
}

func (r *Run) Iterations() int {
	if r.Config.IsPerformance() {
		return 0
	}
	return r.Config.Functional.Iterations
}

func (r *Run) elapsed(at time.Time) int64 {
	return at.Sub(r.StartedAt).Milliseconds()
}

func (r *Run) ResponseKeys() []string {
	keys := make([]string, 0)
	for _, row := range r.Rows {
		if row.KeepsResponse() {
			keys = append(keys, row.ResponseKey)
		}
	}
	return keys
}

// WithoutResponses clears the ResponseKey of rows whose key is listed.
func (r *Run) WithoutResponses(keys []string) Run {
	updated := *r
	rows := make([]RunRowResult, 0, len(r.Rows))
	for _, row := range r.Rows {
		if slices.Contains(keys, row.ResponseKey) {
			row.ResponseKey = ""
		}
		rows = append(rows, row)
	}
	updated.Rows = rows
	return updated
}

// KeptResponseKey is the key of the response kept for the row at index; not found when the row or its response is missing.
func (r *Run) KeptResponseKey(index int) (string, error) {
	for _, row := range r.Rows {
		if row.Index != index {
			continue
		}
		if row.KeepsResponse() {
			return row.ResponseKey, nil
		}
		break
	}
	return "", errs.NotFoundError("no active or stored run with the id "+r.ID, "RUN-01404")
}

func (r *Run) MaskSecrets(mask func(text string) (string, bool)) {
	rows := make([]RunRowResult, 0, len(r.Rows))
	for _, row := range r.Rows {
		rows = append(rows, row.WithSecretsMasked(mask))
	}
	r.Rows = rows
}

func (r *Run) FailedTests() int {
	failed := 0
	for _, row := range r.Rows {
		failed += row.FailedTests()
	}
	return failed
}

func (r *Run) SkippedTests() int {
	skipped := 0
	for _, row := range r.Rows {
		skipped += row.SkippedTests()
	}
	return skipped
}

func (r *Run) BrokenRows() int {
	broken := 0
	for _, row := range r.Rows {
		if row.Breaks() {
			broken++
		}
	}
	return broken
}
