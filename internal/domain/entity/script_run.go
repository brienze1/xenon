package entity

import (
	"maps"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scriptrunstatus"
)

type ScriptRun struct {
	StartedAt   time.Time
	Error       *ScriptError
	NextRequest *NextRequest
	Vault       map[string]string
	Script      ScriptRef
	Status      enums.ScriptRunStatus
	Scopes      ScriptScopes
	Logs        []ScriptLog
	Notices     []ScriptLog
	Nested      []NestedRequest
	Tests       []TestResult
	Request     Request
	Duration    time.Duration
	TestCount   int
}

func NewScriptRun(script ScriptRef, startedAt time.Time) ScriptRun {
	return ScriptRun{Script: script, StartedAt: startedAt, Status: scriptrunstatus.Ok, Logs: []ScriptLog{}, Notices: []ScriptLog{}, Nested: []NestedRequest{}, Tests: []TestResult{}}
}

func (r ScriptRun) WithEffects(effects ScriptEffects) ScriptRun {
	r.Request = effects.Request
	r.NextRequest = effects.NextRequest
	r.Scopes = effects.Scopes
	r.Vault = maps.Clone(effects.Vault)
	r.Logs = append([]ScriptLog{}, effects.Logs...)
	r.Notices = append([]ScriptLog{}, effects.Notices...)
	r.Nested = append([]NestedRequest{}, effects.Nested...)
	r.Tests = effects.TestsOf(r.Script)
	r.TestCount = len(r.Tests)
	return r
}

func (r ScriptRun) Failed(scriptError ScriptError) ScriptRun {
	r.Status = scriptrunstatus.Error
	r.Error = &scriptError
	return r
}

func (r ScriptRun) SkippedRequest() ScriptRun {
	r.Status = scriptrunstatus.Skipped
	return r
}

func (r ScriptRun) Finished(at time.Time) ScriptRun {
	r.Duration = at.Sub(r.StartedAt)
	return r
}

func (r ScriptRun) FailedTests() []TestResult {
	failed := make([]TestResult, 0, len(r.Tests))
	for _, test := range r.Tests {
		if test.IsFailed() {
			failed = append(failed, test)
		}
	}
	return failed
}

func (r ScriptRun) HasFailed() bool {
	return r.Status == scriptrunstatus.Error
}

func (r ScriptRun) HasSkippedRequest() bool {
	return r.Status == scriptrunstatus.Skipped
}

func (r ScriptRun) StopsSend() bool {
	return r.Script.IsPreRequest() && (r.HasFailed() || r.HasSkippedRequest())
}

// ConsoleEntryCount is how many console lines a finished run adds: one per failed test, plus one for its error.
func (r ScriptRun) ConsoleEntryCount() int {
	count := len(r.FailedTests())
	if r.HasFailed() {
		count++
	}
	return count
}

// ConsoleEntries are the console lines a finished run adds: one per failed test, then its error; ids and ats hold one value per entry, in order.
func (r ScriptRun) ConsoleEntries(ids []string, sendId string, ats []time.Time, mask func(string) (string, bool)) []ConsoleEntry {
	failed := r.FailedTests()
	entries := make([]ConsoleEntry, 0, len(failed)+1)
	for index, test := range failed {
		entries = append(entries, NewConsoleTestFailureEntry(ids[index], sendId, test, ats[index]).WithSecretsMasked(mask))
	}
	if r.HasFailed() {
		entries = append(entries, NewConsoleErrorEntry(ids[len(failed)], sendId, *r.Error, ats[len(failed)]).WithSecretsMasked(mask))
	}
	return entries
}
