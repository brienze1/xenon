package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type ActiveRunsPersistence interface {
	usecase.ClaimActiveRun
	usecase.ChangeActiveRun
	usecase.ReleaseActiveRun
	usecase.FindActiveRuns
	usecase.FindActiveRunControls
}
