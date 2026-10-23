package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type RunProgress struct {
	Metrics     *RunPerformanceMetrics `json:"metrics"`
	CurrentName string                 `json:"current_name"`
	CurrentRow  int                    `json:"current_row"`
	Iteration   int                    `json:"iteration"`
	Completed   int                    `json:"completed"`
	Planned     int                    `json:"planned"`
	TestsPassed int                    `json:"tests_passed"`
	TestsFailed int                    `json:"tests_failed"`
	ElapsedMs   int64                  `json:"elapsed_ms"`
}

func NewRunProgress(progress entity.RunProgress) RunProgress {
	return RunProgress{
		Metrics:     NewRunPerformanceMetrics(progress.Metrics),
		CurrentName: progress.CurrentName,
		CurrentRow:  progress.CurrentRow,
		Iteration:   progress.Iteration,
		Completed:   progress.Completed,
		Planned:     progress.Planned,
		TestsPassed: progress.TestsPassed,
		TestsFailed: progress.TestsFailed,
		ElapsedMs:   progress.ElapsedMs,
	}
}
