package dto

import "github.com/brienze1/xenon/internal/domain/enums/teststatus"

type MCPSendTestLine struct {
	Name   string `json:"name"`
	Error  string `json:"error,omitempty"`
	Passed bool   `json:"passed"`
}

func NewMCPSendTestLine(test TestResult) MCPSendTestLine {
	return MCPSendTestLine{Name: test.Name, Passed: teststatus.Passed.Equals(test.Status), Error: test.ErrorMessage}
}
