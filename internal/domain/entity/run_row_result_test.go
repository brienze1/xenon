package entity

import (
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/runrowoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/scripterrorkind"
	"github.com/brienze1/xenon/internal/domain/enums/sendoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
)

func TestNewRunRowOutcome(t *testing.T) {
	failedTest := []TestResult{{Name: "ok", Status: teststatus.Passed}, {Name: "status", Status: teststatus.Failed}}
	scriptError := ScriptError{Kind: scripterrorkind.Runtime, Message: "boom"}
	cases := map[string]struct {
		execution SendExecution
		want      string
	}{
		"skipped":         {SendExecution{Outcome: sendoutcome.SkippedByScript, Tests: failedTest}, runrowoutcome.Skipped.String()},
		"cancelled":       {SendExecution{Outcome: sendoutcome.Cancelled}, runrowoutcome.Cancelled.String()},
		"not sent":        {SendExecution{Outcome: sendoutcome.Error}, runrowoutcome.Error.String()},
		"script stopped":  {SendExecution{Outcome: sendoutcome.StoppedByScriptError}, runrowoutcome.Error.String()},
		"script error":    {SendExecution{Outcome: sendoutcome.Sent, PostRuns: []ScriptRun{{Error: &scriptError}}}, runrowoutcome.Error.String()},
		"failed test":     {SendExecution{Outcome: sendoutcome.Sent, Tests: failedTest}, runrowoutcome.Failed.String()},
		"passed":          {SendExecution{Outcome: sendoutcome.Sent, Tests: failedTest[:1]}, runrowoutcome.Passed.String()},
		"passed no tests": {SendExecution{Outcome: sendoutcome.Sent}, runrowoutcome.Passed.String()},
	}
	for name, testCase := range cases {
		if got := NewRunRowOutcome(testCase.execution); got.String() != testCase.want {
			t.Errorf("%s: got %s, want %s", name, got, testCase.want)
		}
	}
}

func TestRunRowResultWithExecutionReadsTheResponse(t *testing.T) {
	sent := EffectiveRequest{Method: "POST", URL: "http://api/login?key=s3cret", Headers: []KeyValue{NewKeyValue("Authorization", "Bearer s3cret")}}
	response := Response{StatusCode: 201, SizeBytes: 12, Timings: Timings{Total: 42 * time.Millisecond}}
	execution := SendExecution{Outcome: sendoutcome.Sent, SentRequest: &sent, Response: &response, Tests: []TestResult{{Name: "has s3cret", Status: teststatus.Failed, ErrorMessage: "got s3cret"}}}
	row := NewRunRowResult(3, 1, RunPlanRow{RunRow: RunRow{RequestID: "r"}, Name: "Login", FolderPath: []string{"Auth"}}).WithExecution(execution)
	if row.Index != 3 || row.Iteration != 1 || row.Status != 201 || row.DurationMs != 42 || row.SizeBytes != 12 || row.Outcome != runrowoutcome.Failed {
		t.Errorf("row = %+v", row)
	}
	masked := row.WithSecretsMasked(func(text string) (string, bool) {
		masked := strings.ReplaceAll(text, "s3cret", MaskedValue)
		return masked, masked != text
	})
	if masked.URL != "http://api/login?key="+MaskedValue || masked.SentRequest.Headers[0].Value != "Bearer "+MaskedValue || masked.Tests[0].Name != "has "+MaskedValue || masked.Tests[0].ErrorMessage != "got "+MaskedValue {
		t.Errorf("masked = %+v", masked)
	}
}

func TestRunSummaryCountsRowsAndTests(t *testing.T) {
	rows := []RunRowResult{
		{Outcome: runrowoutcome.Passed, Status: 200, DurationMs: 100, Tests: []TestResult{{Status: teststatus.Passed}, {Status: teststatus.Skipped}}},
		{Outcome: runrowoutcome.Failed, Status: 500, DurationMs: 300, Tests: []TestResult{{Status: teststatus.Failed}}},
		{Outcome: runrowoutcome.Skipped},
		{Outcome: runrowoutcome.Error},
		{Outcome: runrowoutcome.Cancelled},
	}
	summary := NewRunSummary(2, rows, 900)
	want := RunSummary{Iterations: 2, TotalRequests: 5, TestsPassed: 1, TestsFailed: 1, TestsSkipped: 1, AvgResponseMs: 200, DurationMs: 900, Passed: 1, Failed: 1, Skipped: 1, Errors: 1, Cancelled: 1}
	if summary != want {
		t.Errorf("summary = %+v, want %+v", summary, want)
	}
}
