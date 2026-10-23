package usecase

import (
	"encoding/xml"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/runrowoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
)

func TestBuildRunJUnitWritesASuitePerRowAndIteration(t *testing.T) {
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
	encoded, err := BuildRunJUnit(run)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	text := string(encoded)
	if !strings.HasPrefix(text, "<?xml") {
		t.Errorf("no header: %s", text)
	}
	var parsed junitSuites
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
