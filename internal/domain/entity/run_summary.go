package entity

import "github.com/brienze1/xenon/internal/domain/enums/runrowoutcome"

type RunSummary struct {
	Iterations    int
	TotalRequests int
	TestsPassed   int
	TestsFailed   int
	TestsSkipped  int
	AvgResponseMs int64
	DurationMs    int64
	Passed        int
	Failed        int
	Skipped       int
	Errors        int
	Cancelled     int
}

func NewRunSummary(iterations int, rows []RunRowResult, durationMs int64) RunSummary {
	summary := RunSummary{Iterations: iterations, TotalRequests: len(rows), DurationMs: durationMs}
	var responseMs int64
	responded := 0
	for _, row := range rows {
		tests := NewTestSummary(row.Tests)
		summary.TestsPassed += tests.Passed
		summary.TestsFailed += tests.Failed
		summary.TestsSkipped += tests.Skipped
		summary = summary.counted(row)
		if row.Status > 0 {
			responseMs += row.DurationMs
			responded++
		}
	}
	if responded > 0 {
		summary.AvgResponseMs = responseMs / int64(responded)
	}
	return summary
}

func (r RunSummary) counted(row RunRowResult) RunSummary {
	switch row.Outcome {
	case runrowoutcome.Passed:
		r.Passed++
	case runrowoutcome.Failed:
		r.Failed++
	case runrowoutcome.Skipped:
		r.Skipped++
	case runrowoutcome.Error:
		r.Errors++
	case runrowoutcome.Cancelled:
		r.Cancelled++
	}
	return r
}

func (r RunSummary) WithPerformance(overall RequestReport) RunSummary {
	r.TotalRequests = overall.Count
	r.AvgResponseMs = int64(overall.AvgMs)
	return r
}
