package entity

import (
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/runrowoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
	"github.com/brienze1/xenon/internal/domain/enums/runtype"
)

func TestRunPauseResumeAndFinish(t *testing.T) {
	started := time.Date(2026, 10, 23, 10, 0, 0, 0, time.UTC)
	run := NewRun("run1", RunTarget{ItemID: "c"}, selectedRunConfig(), 2, started)
	if !run.IsActive() || run.Status != runstatus.Running || run.Progress.Planned != 2 {
		t.Fatalf("run = %+v", run)
	}
	if run.Resume() {
		t.Error("a running run cannot resume")
	}
	if !run.Pause() || !run.IsPaused() || !run.IsActive() {
		t.Errorf("pause: %+v", run)
	}
	if run.Pause() {
		t.Error("a paused run cannot pause again")
	}
	if !run.Resume() || run.Status != runstatus.Running {
		t.Errorf("resume: %+v", run)
	}
	run.Starting(RunPlanRow{Name: "Login", Number: 1}, 0, started.Add(time.Second))
	run.Record(RunRowResult{Outcome: runrowoutcome.Passed}, started.Add(2*time.Second))
	if run.Progress.CurrentName != "Login" || run.Progress.Completed != 1 || run.Progress.ElapsedMs != 2000 {
		t.Errorf("progress = %+v", run.Progress)
	}
	run.Finish(runstatus.Stopped, started.Add(3*time.Second))
	run.Finish(runstatus.Completed, started.Add(4*time.Second))
	if run.Status != runstatus.Stopped || run.IsActive() || run.Summary.TotalRequests != 1 || run.Summary.DurationMs != 3000 || run.Summary.Iterations != 1 {
		t.Errorf("finished = %+v", run)
	}
	if run.Pause() || run.Resume() {
		t.Error("a finished run is terminal")
	}
}

func TestPerformanceRunsNeverPause(t *testing.T) {
	config := selectedRunConfig()
	config.Type = runtype.Performance
	run := NewRun("run1", RunTarget{}, config, 0, time.Now())
	if run.CanPause() || run.Pause() {
		t.Error("a performance run has no pause")
	}
}

func TestRunStartAcknowledgesExpiredTokens(t *testing.T) {
	acknowledgement := RunStart{ConfirmedEnvironmentID: "dev", AcknowledgeMissingSecrets: true}.Acknowledgement()
	if acknowledgement.ConfirmedEnvironmentID != "dev" || !acknowledgement.MissingSecrets || !acknowledgement.ExpiredToken {
		t.Errorf("acknowledgement = %+v", acknowledgement)
	}
}

func TestSendPreflightMergedUnitesTheRows(t *testing.T) {
	environment := Environment{ID: "dev", ConfirmBeforeSend: true}
	first := SendPreflight{MissingSecrets: []string{"b"}}
	second := SendPreflight{ConfirmEnvironment: &environment, MissingSecrets: []string{"a", "b"}, LockedSecrets: []string{"c"}}
	merged := first.Merged(second)
	if merged.ConfirmEnvironment == nil || len(merged.MissingSecrets) != 2 || merged.MissingSecrets[0] != "a" || len(merged.LockedSecrets) != 1 {
		t.Errorf("merged = %+v", merged)
	}
}
