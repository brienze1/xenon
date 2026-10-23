package persistence

import (
	"bytes"
	"context"
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/consoleentrykind"
	"github.com/brienze1/xenon/internal/domain/enums/consolelevel"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/runexportformat"
	"github.com/brienze1/xenon/internal/domain/enums/runrowoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
	"github.com/brienze1/xenon/internal/domain/enums/runtype"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptownerkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
)

const goldenRunJunit = `<?xml version="1.0" encoding="UTF-8"?>
<testsuites name="Shop &amp; &lt;co&gt;" time="1.500" tests="6" failures="1" errors="3" skipped="1">
  <testsuite name="Login [iteration 1]" timestamp="2026-10-23T11:00:00" time="0.120" tests="3" failures="1" errors="0" skipped="1">
    <testcase name="ok" classname="Login" time="0.000"></testcase>
    <testcase name="status" classname="Login" time="0.000">
      <failure message="expected 200 &#34;x&#34;">expected 200 &#34;x&#34;</failure>
    </testcase>
    <testcase name="later" classname="Login" time="0.000">
      <skipped></skipped>
    </testcase>
  </testsuite>
  <testsuite name="Orders/Create [iteration 2]" timestamp="2026-10-23T11:00:00" time="0.000" tests="1" failures="0" errors="1" skipped="0">
    <testcase name="request" classname="Create" time="0.000">
      <error message="connection refused">connection refused</error>
    </testcase>
  </testsuite>
  <testsuite name="Orders/Deep/Get [iteration 1]" timestamp="2026-10-23T11:00:00" time="0.000" tests="1" failures="0" errors="1" skipped="0">
    <testcase name="request" classname="Get" time="0.000">
      <error message="boom">boom</error>
    </testcase>
  </testsuite>
  <testsuite name="Logout [iteration 1]" timestamp="2026-10-23T11:00:00" time="0.000" tests="1" failures="0" errors="1" skipped="0">
    <testcase name="request" classname="Logout" time="0.000">
      <error message="cancelled">cancelled</error>
    </testcase>
  </testsuite>
  <testsuite name="Empty [iteration 1]" timestamp="2026-10-23T11:00:00" time="0.000" tests="0" failures="0" errors="0" skipped="0"></testsuite>
</testsuites>
`

func exportedRun(t *testing.T, run entity.Run, format string) []byte {
	t.Helper()
	target := filepath.Join(t.TempDir(), "out", "run."+format)
	if err := NewExportFilePersistence().SaveRunExport(context.Background(), target, run, runexportformat.GetRunExportFormat(format)); err != nil {
		t.Fatalf("SaveRunExport() error = %v", err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestRunJunitExportWritesASuitePerRowAndIteration(t *testing.T) {
	run := entity.Run{
		Target:    entity.RunTarget{Name: "Shop"},
		StartedAt: time.Date(2026, 10, 23, 12, 0, 0, 0, time.UTC),
		Summary:   entity.RunSummary{DurationMs: 1500},
		Rows: []entity.RunRowResult{
			{Name: "Login", FolderPath: []string{}, Outcome: runrowoutcome.Failed, DurationMs: 120, Tests: []entity.TestResult{
				{Name: "ok", Status: teststatus.Passed},
				{Name: "status", Status: teststatus.Failed, ErrorMessage: "expected 200"},
				{Name: "later", Status: teststatus.Skipped},
			}},
			{Name: "Create", FolderPath: []string{"Orders"}, Iteration: 1, Outcome: runrowoutcome.Error, Error: "connection refused", Tests: []entity.TestResult{}},
			{Name: "Get", FolderPath: []string{"Orders", "Deep"}, Outcome: runrowoutcome.Error, ScriptErrors: []entity.ScriptError{{Message: "boom"}}},
			{Name: "Logout", Outcome: runrowoutcome.Cancelled},
		},
	}
	encoded := exportedRun(t, run, "junit")
	text := string(encoded)
	if !strings.HasPrefix(text, "<?xml") {
		t.Errorf("no header: %s", text)
	}
	var parsed model.RunJunitSuites
	if err := xml.Unmarshal(encoded, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Tests != 6 || parsed.Failures != 1 || parsed.Errors != 3 || parsed.Skipped != 1 || parsed.Time != "1.500" {
		t.Errorf("totals = %+v", parsed)
	}
	names := []string{parsed.Suites[0].Name, parsed.Suites[1].Name, parsed.Suites[2].Name}
	if names[0] != "Login [iteration 1]" || names[1] != "Orders/Create [iteration 2]" || names[2] != "Orders/Deep/Get [iteration 1]" {
		t.Errorf("names = %v", names)
	}
	login := parsed.Suites[0]
	if login.Tests != 3 || login.Time != "0.120" || login.Cases[1].Failure == nil || login.Cases[1].Failure.Message != "expected 200" || login.Cases[2].Skipped == nil || login.Cases[0].Classname != "Login" {
		t.Errorf("login = %+v", login)
	}
	if parsed.Suites[1].Cases[0].Error == nil || parsed.Suites[1].Cases[0].Error.Message != "connection refused" || parsed.Suites[1].Cases[0].Name != "request" {
		t.Errorf("create = %+v", parsed.Suites[1])
	}
	if parsed.Suites[2].Cases[0].Error.Message != "boom" || parsed.Suites[3].Cases[0].Error.Message != "cancelled" {
		t.Errorf("break messages = %+v, %+v", parsed.Suites[2], parsed.Suites[3])
	}
}

func TestRunJunitExportBytesAreStable(t *testing.T) {
	run := entity.Run{
		Target:    entity.RunTarget{Name: "Shop & <co>"},
		StartedAt: time.Date(2026, 10, 23, 12, 0, 0, 0, time.FixedZone("x", 3600)),
		Summary:   entity.RunSummary{DurationMs: 1500},
		Rows: []entity.RunRowResult{
			{Name: "Login", FolderPath: []string{}, Outcome: runrowoutcome.Failed, DurationMs: 120, Tests: []entity.TestResult{
				{Name: "ok", Status: teststatus.Passed},
				{Name: "status", Status: teststatus.Failed, ErrorMessage: "expected 200 \"x\""},
				{Name: "later", Status: teststatus.Skipped},
			}},
			{Name: "Create", FolderPath: []string{"Orders"}, Iteration: 1, Outcome: runrowoutcome.Error, Error: "connection refused", Tests: []entity.TestResult{}},
			{Name: "Get", FolderPath: []string{"Orders", "Deep"}, Outcome: runrowoutcome.Error, ScriptErrors: []entity.ScriptError{{Message: "boom"}}},
			{Name: "Logout", Outcome: runrowoutcome.Cancelled},
			{Name: "Empty", Outcome: runrowoutcome.Passed},
		},
	}
	if got := string(exportedRun(t, run, "junit")); got != goldenRunJunit {
		t.Errorf("junit = %s", got)
	}
}

func TestRunJsonExportIsTheStoredRunFile(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XENON_DATA_DIR", dataDir)
	ctx := context.Background()
	at := time.Date(2026, 10, 23, 12, 0, 0, 123456789, time.FixedZone("x", 3600))
	functional := entity.Run{
		ID: "run-1", Status: runstatus.Completed, StartedAt: at, FinishedAt: at.Add(time.Second), EnvironmentName: "Dev", DataFileName: "data.csv",
		Target: entity.RunTarget{Kind: itemkind.Collection, LocalID: "ws-1", ItemID: "shop", Name: "Shop"},
		Config: entity.NewRunConfig(),
		Rows: []entity.RunRowResult{{
			Index: 0, Iteration: 1, RequestID: "r1", AliasID: "a1", AliasName: "Alias", Name: "Login", Method: "POST", URL: "http://api/login", Status: 201, DurationMs: 42, SizeBytes: 12,
			Outcome: runrowoutcome.Failed, ResponseKey: "run:run-1:0", FolderPath: []string{"Auth"},
			SentRequest:  entity.SentRequest{Method: "POST", URL: "http://api/login", Body: "{}", Headers: []entity.KeyValue{entity.NewKeyValue("Accept", "*/*")}},
			Tests:        []entity.TestResult{{Name: "status", Status: teststatus.Failed, ErrorMessage: "expected 200", Script: entity.ScriptRef{Phase: scriptphase.PostResponse, OwnerKind: scriptownerkind.Request, ItemID: "r1", OwnerName: "Login"}}},
			ScriptErrors: []entity.ScriptError{{Script: entity.ScriptRef{Phase: scriptphase.PostResponse, OwnerKind: scriptownerkind.Request, ItemID: "r1", OwnerName: "Login", Path: "login.yaml"}, Kind: scripterrorkind.Runtime, Message: "boom", Line: 3, Column: 7}},
			Console:      []entity.ConsoleEntry{{CreatedAt: at, Kind: consoleentrykind.Log, Level: consolelevel.Info, ID: "c1", Summary: "hello", Arguments: []string{"hello"}, Values: []any{"hello", 1.5, true}, Script: &entity.ScriptRef{Phase: scriptphase.PreRequest, OwnerKind: scriptownerkind.Collection, ItemID: "shop", OwnerName: "Shop"}}},
		}},
		Notices:  []string{"a notice"},
		Progress: entity.RunProgress{Completed: 1, Planned: 1, TestsFailed: 1, ElapsedMs: 1000},
		Summary:  entity.RunSummary{Iterations: 1, TotalRequests: 1, TestsFailed: 1, AvgResponseMs: 42, DurationMs: 1000, Failed: 1},
	}
	performance := functional
	performance.ID = "run-2"
	performance.Config.Type = runtype.Performance
	performance.Rows = []entity.RunRowResult{}
	performance.Progress.Metrics = &entity.PerformanceMetrics{ActiveUsers: 2, Requests: 10, AvgMs: 12.5, Throughput: 3.25, ErrorRate: 0.1}
	performance.Performance = &entity.PerformanceReport{
		Overall:    entity.RequestReport{StatusCounts: map[int]int{200: 9, 500: 1}, Name: "All", Count: 10, AvgMs: 12.5, P90Ms: 20, P95Ms: 22, P99Ms: 30, MinMs: 1, MaxMs: 31, Throughput: 3.25, ErrorRate: 0.1},
		PerRequest: []entity.RequestReport{{StatusCounts: map[int]int{200: 9, 500: 1}, RequestID: "r1", Name: "Login", Count: 10, AvgMs: 12.5}},
		Series:     []entity.PerformanceSecond{{Second: 1, ActiveUsers: 2, Requests: 10, Errors: 1, AvgMs: 12.5}},
	}
	runs := NewRunPersistence()
	for _, run := range []entity.Run{functional, performance} {
		if err := runs.SaveRun(ctx, run); err != nil {
			t.Fatal(err)
		}
		stored, err := os.ReadFile(filepath.Join(runsFolder("ws-1", "shop"), run.ID+runFileSuffix))
		if err != nil {
			t.Fatal(err)
		}
		found, err := runs.FindRun(ctx, "ws-1", run.ID)
		if err != nil {
			t.Fatal(err)
		}
		if exported := exportedRun(t, *found, "json"); !bytes.Equal(exported, stored) {
			t.Errorf("%s export differs from the stored run:\n%s\n---\n%s", run.ID, exported, stored)
		}
	}
}
