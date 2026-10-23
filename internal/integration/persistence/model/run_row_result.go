package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/runrowoutcome"
)

type RunRowResult struct {
	Outcome      string            `json:"outcome"`
	RequestId    string            `json:"request_id"`
	AliasId      string            `json:"alias_id"`
	AliasName    string            `json:"alias_name"`
	Name         string            `json:"name"`
	Method       string            `json:"method"`
	Url          string            `json:"url"`
	Error        string            `json:"error"`
	ResponseKey  string            `json:"response_key"`
	SentRequest  RunSentRequest    `json:"sent_request"`
	FolderPath   []string          `json:"folder_path"`
	Tests        []RunTestResult   `json:"tests"`
	ScriptErrors []RunScriptError  `json:"script_errors"`
	Console      []RunConsoleEntry `json:"console"`
	DurationMs   int64             `json:"duration_ms"`
	SizeBytes    int64             `json:"size_bytes"`
	Index        int               `json:"index"`
	Iteration    int               `json:"iteration"`
	Status       int               `json:"status"`
}

func NewRunRowResult(row entity.RunRowResult) RunRowResult {
	tests := make([]RunTestResult, 0, len(row.Tests))
	for _, test := range row.Tests {
		tests = append(tests, NewRunTestResult(test))
	}
	scriptErrors := make([]RunScriptError, 0, len(row.ScriptErrors))
	for _, scriptError := range row.ScriptErrors {
		scriptErrors = append(scriptErrors, NewRunScriptError(scriptError))
	}
	console := make([]RunConsoleEntry, 0, len(row.Console))
	for _, entry := range row.Console {
		console = append(console, NewRunConsoleEntry(entry))
	}
	return RunRowResult{
		Outcome:      row.Outcome.String(),
		RequestId:    row.RequestID,
		AliasId:      row.AliasID,
		AliasName:    row.AliasName,
		Name:         row.Name,
		Method:       row.Method,
		Url:          row.URL,
		Error:        row.Error,
		ResponseKey:  row.ResponseKey,
		SentRequest:  NewRunSentRequest(row.SentRequest),
		FolderPath:   append([]string{}, row.FolderPath...),
		Tests:        tests,
		ScriptErrors: scriptErrors,
		Console:      console,
		DurationMs:   row.DurationMs,
		SizeBytes:    row.SizeBytes,
		Index:        row.Index,
		Iteration:    row.Iteration,
		Status:       row.Status,
	}
}

func (r *RunRowResult) ToEntity() entity.RunRowResult {
	tests := make([]entity.TestResult, 0, len(r.Tests))
	for _, test := range r.Tests {
		tests = append(tests, test.ToEntity())
	}
	scriptErrors := make([]entity.ScriptError, 0, len(r.ScriptErrors))
	for _, scriptError := range r.ScriptErrors {
		scriptErrors = append(scriptErrors, scriptError.ToEntity())
	}
	console := make([]entity.ConsoleEntry, 0, len(r.Console))
	for _, entry := range r.Console {
		console = append(console, entry.ToEntity())
	}
	return entity.RunRowResult{
		Outcome:      runrowoutcome.GetRunRowOutcome(r.Outcome),
		RequestID:    r.RequestId,
		AliasID:      r.AliasId,
		AliasName:    r.AliasName,
		Name:         r.Name,
		Method:       r.Method,
		URL:          r.Url,
		Error:        r.Error,
		ResponseKey:  r.ResponseKey,
		SentRequest:  r.SentRequest.ToEntity(),
		FolderPath:   append([]string{}, r.FolderPath...),
		Tests:        tests,
		ScriptErrors: scriptErrors,
		Console:      console,
		DurationMs:   r.DurationMs,
		SizeBytes:    r.SizeBytes,
		Index:        r.Index,
		Iteration:    r.Iteration,
		Status:       r.Status,
	}
}
