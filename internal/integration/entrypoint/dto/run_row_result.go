package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunRowResult struct {
	RequestId       string          `json:"request_id"`
	AliasId         string          `json:"alias_id"`
	AliasName       string          `json:"alias_name"`
	Name            string          `json:"name"`
	Method          string          `json:"method"`
	Url             string          `json:"url"`
	Outcome         string          `json:"outcome"`
	Error           string          `json:"error"`
	ResponseBodyUrl string          `json:"response_body_url"`
	SentRequest     RunSentRequest  `json:"sent_request"`
	FolderPath      []string        `json:"folder_path"`
	Tests           []RunTestResult `json:"tests"`
	ScriptErrors    []ScriptError   `json:"script_errors"`
	Console         []ConsoleEntry  `json:"console"`
	DurationMs      int64           `json:"duration_ms"`
	SizeBytes       int64           `json:"size_bytes"`
	Index           int             `json:"index"`
	Iteration       int             `json:"iteration"`
	Status          int             `json:"status"`
}

func NewRunRowResult(row entity.RunRowResult) RunRowResult {
	tests := make([]RunTestResult, 0, len(row.Tests))
	for _, test := range row.Tests {
		tests = append(tests, NewRunTestResult(test))
	}
	scriptErrors := make([]ScriptError, 0, len(row.ScriptErrors))
	for _, scriptError := range row.ScriptErrors {
		scriptErrors = append(scriptErrors, NewScriptError(scriptError))
	}
	result := RunRowResult{
		Index:        row.Index,
		Iteration:    row.Iteration,
		RequestId:    row.RequestID,
		AliasId:      row.AliasID,
		AliasName:    row.AliasName,
		Name:         row.Name,
		Method:       row.Method,
		FolderPath:   append([]string{}, row.FolderPath...),
		Url:          row.URL,
		Status:       row.Status,
		DurationMs:   row.DurationMs,
		SizeBytes:    row.SizeBytes,
		Outcome:      row.Outcome.String(),
		Tests:        tests,
		ScriptErrors: scriptErrors,
		Console:      NewConsoleEntries(row.Console),
		SentRequest:  NewRunSentRequest(row.SentRequest),
		Error:        row.Error,
	}
	if row.ResponseKey != "" {
		result.ResponseBodyUrl = ResponseBodyRoute + "/" + row.ResponseKey
	}
	return result
}
