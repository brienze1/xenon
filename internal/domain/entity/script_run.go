package entity

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/scriptrunstatus"
)

type ScriptRun struct {
	StartedAt      time.Time
	Error          *ScriptError
	LocalVariables map[string]any
	Script         ScriptRef
	Status         enums.ScriptRunStatus
	Logs           []ScriptLog
	Nested         []NestedRequest
	Request        Request
	Duration       time.Duration
	TestCount      int
}

func NewScriptRun(script ScriptRef, startedAt time.Time) ScriptRun {
	return ScriptRun{Script: script, StartedAt: startedAt, Status: scriptrunstatus.Ok, Logs: []ScriptLog{}, Nested: []NestedRequest{}}
}

func (r ScriptRun) WithEffects(request Request, localVariables map[string]any, logs []ScriptLog, nested []NestedRequest) ScriptRun {
	r.Request = request
	r.LocalVariables = localVariables
	r.Logs = append([]ScriptLog{}, logs...)
	r.Nested = append([]NestedRequest{}, nested...)
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
