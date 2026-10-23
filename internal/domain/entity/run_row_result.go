package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runrowoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/sendoutcome"
)

type RunRowResult struct {
	Outcome      enums.RunRowOutcome
	RequestID    string
	AliasID      string
	AliasName    string
	Name         string
	Method       string
	URL          string
	Error        string
	ResponseKey  string
	SentRequest  SentRequest
	FolderPath   []string
	Tests        []TestResult
	ScriptErrors []ScriptError
	Console      []ConsoleEntry
	DurationMs   int64
	SizeBytes    int64
	Index        int
	Iteration    int
	Status       int
}

func NewRunRowResult(index, iteration int, row RunPlanRow) RunRowResult {
	return RunRowResult{
		Index:        index,
		Iteration:    iteration,
		RequestID:    row.RequestID,
		AliasID:      row.AliasID,
		Name:         row.Name,
		Method:       row.Method,
		FolderPath:   append([]string{}, row.FolderPath...),
		Tests:        []TestResult{},
		ScriptErrors: []ScriptError{},
		Console:      []ConsoleEntry{},
	}
}

// NewRunRowOutcome derives a row's outcome: skipped, cancelled, error (not sent or a script error), failed (a failed test) or passed.
func NewRunRowOutcome(execution SendExecution) enums.RunRowOutcome {
	switch {
	case execution.Outcome == sendoutcome.SkippedByScript:
		return runrowoutcome.Skipped
	case execution.Outcome == sendoutcome.Cancelled:
		return runrowoutcome.Cancelled
	case execution.Outcome != sendoutcome.Sent || len(execution.ScriptErrors()) > 0:
		return runrowoutcome.Error
	case execution.TestSummary().Failed > 0:
		return runrowoutcome.Failed
	default:
		return runrowoutcome.Passed
	}
}

func (r RunRowResult) ForRequest(request Request, alias RequestAlias) RunRowResult {
	r.Name = request.Name
	r.Method = request.NormalizedMethod()
	r.URL = request.URL
	r.AliasName = alias.Name
	return r
}

func (r RunRowResult) WithExecution(execution SendExecution) RunRowResult {
	r.Outcome = NewRunRowOutcome(execution)
	r.Tests = append([]TestResult{}, execution.Tests...)
	r.ScriptErrors = execution.ScriptErrors()
	if execution.SentRequest != nil {
		r.SentRequest = NewSentRequest(*execution.SentRequest)
		r.URL = execution.SentRequest.URL
		r.Method = execution.SentRequest.Method
	}
	if execution.Response != nil {
		r.Status = execution.Response.StatusCode
		r.DurationMs = execution.Response.Timings.Total.Milliseconds()
		r.SizeBytes = execution.Response.SizeBytes
		if execution.Response.Error != nil {
			r.Status = 0
			r.Error = execution.Response.Error.Message
		}
	}
	return r
}

func (r RunRowResult) Failed(message string) RunRowResult {
	r.Outcome = runrowoutcome.Error
	r.Error = message
	return r
}

func (r RunRowResult) Cancelled() RunRowResult {
	r.Outcome = runrowoutcome.Cancelled
	return r
}

func (r RunRowResult) WithConsole(entries []ConsoleEntry) RunRowResult {
	r.Console = append([]ConsoleEntry{}, entries...)
	return r
}

func (r RunRowResult) WithSecretsMasked(mask func(text string) (string, bool)) RunRowResult {
	masker := secretMasker{mask: mask}
	r.URL = masker.text(r.URL)
	r.Error = masker.text(r.Error)
	r.SentRequest = r.SentRequest.WithSecretsMasked(mask)
	tests := make([]TestResult, 0, len(r.Tests))
	for _, test := range r.Tests {
		test.Name = masker.text(test.Name)
		test.ErrorMessage = masker.text(test.ErrorMessage)
		tests = append(tests, test)
	}
	r.Tests = tests
	scriptErrors := make([]ScriptError, 0, len(r.ScriptErrors))
	for _, scriptError := range r.ScriptErrors {
		scriptErrors = append(scriptErrors, scriptError.WithMessage(masker.text(scriptError.Message)))
	}
	r.ScriptErrors = scriptErrors
	console := make([]ConsoleEntry, 0, len(r.Console))
	for _, entry := range r.Console {
		console = append(console, entry.WithSecretsMasked(mask))
	}
	r.Console = console
	return r
}
