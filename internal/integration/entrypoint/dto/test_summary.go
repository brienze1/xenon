package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type TestSummary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

func NewTestSummary(summary entity.TestSummary) TestSummary {
	return TestSummary{Total: summary.Total, Passed: summary.Passed, Failed: summary.Failed, Skipped: summary.Skipped}
}
