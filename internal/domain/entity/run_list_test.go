package entity

import (
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runrowoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
	"github.com/brienze1/xenon/internal/domain/enums/runtype"
)

func listedRun(id, itemId string, startedAt time.Time, status enums.RunStatus) Run {
	run := NewRun(id, RunTarget{LocalID: "L1", ItemID: itemId}, selectedRunConfig(), 2, startedAt)
	run.Status = status
	return run
}

func runIds(runs []Run) []string {
	ids := make([]string, 0, len(runs))
	for _, run := range runs {
		ids = append(ids, run.ID)
	}
	return ids
}

func TestListRunsMergesTheTargetsActiveRunWithItsStoredRuns(t *testing.T) {
	at := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	stored := []Run{
		listedRun("stored2", "shop", at.Add(2*time.Minute), runstatus.Completed),
		listedRun("stored1", "shop", at.Add(time.Minute), runstatus.Stopped),
	}
	finishing := listedRun("stored2", "shop", at.Add(2*time.Minute), runstatus.Running)
	cases := []struct {
		name   string
		active []Run
		stored []Run
		want   []string
	}{
		{name: "no runs", active: nil, stored: nil, want: []string{}},
		{name: "stored only", active: nil, stored: stored, want: []string{"stored2", "stored1"}},
		{name: "the active run comes first", active: []Run{listedRun("active", "shop", at.Add(3*time.Minute), runstatus.Running)}, stored: stored, want: []string{"active", "stored2", "stored1"}},
		{name: "a paused run is listed", active: []Run{listedRun("paused", "shop", at.Add(3*time.Minute), runstatus.Paused)}, stored: stored, want: []string{"paused", "stored2", "stored1"}},
		{name: "a run both stored and active is listed once", active: []Run{finishing}, stored: stored, want: []string{"stored2", "stored1"}},
		{name: "active runs of other targets are left out", active: []Run{listedRun("orders", "orders", at.Add(3*time.Minute), runstatus.Running)}, stored: stored, want: []string{"stored2", "stored1"}},
		{name: "equal start times keep the stored order", active: nil, stored: []Run{listedRun("b", "shop", at, runstatus.Completed), listedRun("a", "shop", at, runstatus.Completed)}, want: []string{"b", "a"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := ListRuns("shop", test.active, test.stored)
			if got == nil {
				t.Fatal("ListRuns returned nil")
			}
			ids := runIds(got)
			if len(ids) != len(test.want) {
				t.Fatalf("ids = %v, want %v", ids, test.want)
			}
			for index := range ids {
				if ids[index] != test.want[index] {
					t.Fatalf("ids = %v, want %v", ids, test.want)
				}
			}
		})
	}
	merged := ListRuns("shop", []Run{finishing}, stored)
	if merged[0].Status != runstatus.Completed {
		t.Errorf("the stored copy wins, got status %s", merged[0].Status)
	}
}

func TestListedRunCarriesTheSummarySoFar(t *testing.T) {
	at := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	run := NewRun("run1", RunTarget{ItemID: "shop"}, selectedRunConfig(), 3, at)
	run.Record(RunRowResult{Outcome: runrowoutcome.Passed}, at.Add(time.Second))
	run.Record(RunRowResult{Outcome: runrowoutcome.Failed}, at.Add(2*time.Second))

	listed := run.Listed()
	if listed.Summary.TotalRequests != 2 || listed.Summary.DurationMs != 2000 || listed.Summary.Iterations != 1 {
		t.Errorf("summary so far = %+v", listed.Summary)
	}
	if run.Summary.TotalRequests != 0 {
		t.Errorf("Listed changed the run itself: %+v", run.Summary)
	}

	run.Finish(runstatus.Completed, at.Add(5*time.Second))
	if finished := run.Listed(); finished.Summary != run.Summary {
		t.Errorf("a finished run lists as stored: %+v, want %+v", finished.Summary, run.Summary)
	}
}

func TestListedPerformanceRunCarriesItsPerformanceFiguresSoFar(t *testing.T) {
	config := selectedRunConfig()
	config.Type = runtype.Performance
	at := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	run := NewRun("perf", RunTarget{ItemID: "shop"}, config, 0, at)
	if listed := run.Listed(); listed.Summary.TotalRequests != 0 || listed.Summary.AvgResponseMs != 0 {
		t.Errorf("without a report = %+v", listed.Summary)
	}
	run.Measure(NewPerformanceReport(RequestReport{Count: 40, AvgMs: 12.5}, nil, nil), PerformanceMetrics{}, at.Add(time.Second))
	if listed := run.Listed(); listed.Summary.TotalRequests != 40 || listed.Summary.AvgResponseMs != 12 {
		t.Errorf("with a report = %+v", listed.Summary)
	}
}
