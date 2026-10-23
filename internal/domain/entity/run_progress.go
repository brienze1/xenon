package entity

type RunProgress struct {
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
