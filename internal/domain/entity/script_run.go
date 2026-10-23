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
	r.Tests = make([]TestResult, 0, len(effects.Tests))
	for _, test := range effects.Tests {
		test.Script = r.Script
		r.Tests = append(r.Tests, test)
	}
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

func (r ScriptRun) HasFailed() bool {
	return r.Status == scriptrunstatus.Error
}

func (r ScriptRun) HasSkippedRequest() bool {
	return r.Status == scriptrunstatus.Skipped
}

func (r ScriptRun) StopsSend() bool {
	return r.Script.IsPreRequest() && (r.HasFailed() || r.HasSkippedRequest())
}
