package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type WorkspaceWatcherWebservice interface {
	usecase.WatchWorkspace
	Consume(consumer func(localId, path string))
}
