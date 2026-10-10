package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
)

type TestResult struct {
	Script       ScriptRef
	Status       enums.TestStatus
	Name         string
	ErrorMessage string
	Index        int
	Async        bool
}

func (t TestResult) IsPassed() bool {
	return t.Status == teststatus.Passed
}

func (t TestResult) IsFailed() bool {
	return t.Status == teststatus.Failed
}

func (t TestResult) IsSkipped() bool {
	return t.Status == teststatus.Skipped
}
