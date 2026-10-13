package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type TestResult struct {
	Name         string    `json:"name"`
	Status       string    `json:"status"`
	ErrorMessage string    `json:"error_message,omitempty"`
	Script       ScriptRef `json:"script"`
	Index        int       `json:"index"`
	Async        bool      `json:"async"`
}

func NewTestResult(test entity.TestResult) TestResult {
	return TestResult{Index: test.Index, Name: test.Name, Status: test.Status.String(), ErrorMessage: test.ErrorMessage, Script: NewScriptRef(test.Script), Async: test.Async}
}
