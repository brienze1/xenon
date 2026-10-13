package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type WorkspaceWatcherWebservice interface {
	usecase.WatchWorkspace
	usecase.UnwatchWorkspace
}
