package model

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
)

type RunTestResult struct {
	Name         string       `json:"name"`
	Status       string       `json:"status"`
	ErrorMessage string       `json:"error_message"`
	Script       RunScriptRef `json:"script"`
	Index        int          `json:"index"`
	Async        bool         `json:"async"`
}

func NewRunTestResult(test entity.TestResult) RunTestResult {
	return RunTestResult{Name: test.Name, Status: test.Status.String(), ErrorMessage: test.ErrorMessage, Script: NewRunScriptRef(test.Script), Index: test.Index, Async: test.Async}
}

func (r *RunTestResult) ToEntity() entity.TestResult {
	return entity.TestResult{Name: r.Name, Status: teststatus.GetTestStatus(r.Status), ErrorMessage: r.ErrorMessage, Script: r.Script.ToEntity(), Index: r.Index, Async: r.Async}
}
