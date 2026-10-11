package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type UIStatePersistence interface {
	usecase.FindUIState
	usecase.SaveUIState
	usecase.DeleteUIState
	usecase.FindDigestState
	usecase.SaveDigestState
}
