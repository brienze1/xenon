package entity

import (
	"reflect"
	"testing"
)

func TestStaleRunnerPreferencesListsRemovedTargets(t *testing.T) {
	preferences := map[string]RunnerPreferences{"shop": {}, "gone": {}, "orders": {}, "also-gone": {}}
	if stale := StaleRunnerPreferences(preferences, []string{"shop", "orders"}); !reflect.DeepEqual(stale, []string{"also-gone", "gone"}) {
		t.Errorf("stale = %v", stale)
	}
}

func TestActiveRunOfFindsTheRunOfATarget(t *testing.T) {
	active := []ActiveRun{{Run: Run{ID: "one", Target: RunTarget{ItemID: "shop"}}}, {Run: Run{ID: "two", Target: RunTarget{ItemID: "orders"}}}}
	if run, found := ActiveRunOf(active, "orders"); !found || run.ID != "two" {
		t.Errorf("run = %+v, %v", run, found)
	}
	if _, found := ActiveRunOf(active, "flow"); found {
		t.Error("a target without a run has none")
	}
}

func TestRunConfigDataFilesAndCompleted(t *testing.T) {
	config := RunConfig{Functional: FunctionalSettings{DataFile: "/a.csv"}, Performance: PerformanceSettings{DataFile: "/b.json"}}
	if !reflect.DeepEqual(config.DataFiles(), []string{"/a.csv", "/b.json"}) || config.DataFile() != "/a.csv" {
		t.Fatalf("data files = %v", config.DataFiles())
	}
	cleared := config.WithoutDataFile("/a.csv")
	if cleared.Functional.DataFile != "" || cleared.Performance.DataFile != "/b.json" {
		t.Errorf("cleared = %+v", cleared)
	}
	completed := RunConfig{Functional: FunctionalSettings{DelayMs: 5}}.Completed()
	if completed.Type != NewRunConfig().Type || completed.Functional.Iterations != 1 || completed.Functional.DelayMs != 5 || completed.Performance.VirtualUsers != DefaultVirtualUsers || completed.Rows == nil {
		t.Errorf("completed = %+v", completed)
	}
	rows := NewRunConfig().WithRows([]RunPlanRow{{RunRow: RunRow{RequestID: "a", Selected: true}}})
	if len(rows.Rows) != 1 || rows.Rows[0].RequestID != "a" {
		t.Errorf("rows = %+v", rows.Rows)
	}
}

func TestFunctionalSettingsDelayAndStopOnError(t *testing.T) {
	settings := FunctionalSettings{DelayMs: 300, StopOnError: true}
	if settings.DelayBefore(0) != 0 || settings.DelayBefore(1) != settings.Delay() {
		t.Error("no delay before the first row, the delay before later ones")
	}
	failed := RunRowResult{Outcome: "error"}
	if !settings.StopsAfter(failed) || (FunctionalSettings{}).StopsAfter(failed) || settings.StopsAfter(RunRowResult{Outcome: "failed"}) {
		t.Error("only an error row stops a run with stop on error")
	}
}
