package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunTestResult struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error"`
}

func NewRunTestResult(test entity.TestResult) RunTestResult {
	return RunTestResult{Name: test.Name, Status: test.Status.String(), Error: test.ErrorMessage}
}
