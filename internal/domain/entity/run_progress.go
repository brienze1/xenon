package entity

type RunProgress struct {
	Metrics     *PerformanceMetrics
	CurrentName string
	CurrentRow  int
	Iteration   int
	Completed   int
	Planned     int
	TestsPassed int
	TestsFailed int
	ElapsedMs   int64
}

func NewRunProgress(planned int) RunProgress {
	return RunProgress{Planned: planned}
}

func (r RunProgress) Starting(row RunPlanRow, iteration int) RunProgress {
	r.CurrentRow = row.Number
	r.CurrentName = row.Name
	r.Iteration = iteration
	return r
}

func (r RunProgress) After(row RunRowResult) RunProgress {
	tests := NewTestSummary(row.Tests)
	r.Completed++
	r.TestsPassed += tests.Passed
	r.TestsFailed += tests.Failed
	return r
}

func (r RunProgress) Measured(metrics PerformanceMetrics) RunProgress {
	r.Metrics = &metrics
	r.Completed = metrics.Requests
	r.Planned = metrics.Requests
	return r
}
