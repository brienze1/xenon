package entity

type TestSummary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
}

func NewTestSummary(tests []TestResult) TestSummary {
	summary := TestSummary{Total: len(tests)}
	for _, test := range tests {
		switch {
		case test.IsPassed():
			summary.Passed++
		case test.IsFailed():
			summary.Failed++
		case test.IsSkipped():
			summary.Skipped++
		}
	}
	return summary
}
