package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type SyncSessionPersistence interface {
	usecase.SaveSyncSession
	usecase.FindSyncSession
}
