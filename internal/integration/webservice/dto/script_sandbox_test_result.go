package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
)

type ScriptSandboxTestResult struct {
	Error   *string `json:"error"`
	Name    string  `json:"name"`
	Index   int     `json:"index"`
	Passed  bool    `json:"passed"`
	Skipped bool    `json:"skipped"`
	Async   bool    `json:"async"`
}

func (s ScriptSandboxTestResult) ToEntity() entity.TestResult {
	result := entity.TestResult{Name: s.Name, Index: s.Index, Async: s.Async, Status: teststatus.Passed}
	switch {
	case s.Skipped:
		result.Status = teststatus.Skipped
	case !s.Passed:
		result.Status = teststatus.Failed
		if s.Error != nil {
			result.ErrorMessage = *s.Error
		}
	}
	return result
}
