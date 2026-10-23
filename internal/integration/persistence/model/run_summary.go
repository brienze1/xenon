package model

import "github.com/brienze1/xenon/internal/domain/entity"

type RunSummary struct {
	Iterations    int   `json:"iterations"`
	TotalRequests int   `json:"total_requests"`
	TestsPassed   int   `json:"tests_passed"`
	TestsFailed   int   `json:"tests_failed"`
	TestsSkipped  int   `json:"tests_skipped"`
	AvgResponseMs int64 `json:"avg_response_ms"`
	DurationMs    int64 `json:"duration_ms"`
	Passed        int   `json:"passed"`
	Failed        int   `json:"failed"`
	Skipped       int   `json:"skipped"`
	Errors        int   `json:"errors"`
	Cancelled     int   `json:"cancelled"`
}

func NewRunSummary(summary entity.RunSummary) RunSummary {
	return RunSummary(summary)
}

func (r *RunSummary) ToEntity() entity.RunSummary {
	return entity.RunSummary(*r)
}
