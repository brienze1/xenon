package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type HistoryPersistence interface {
	usecase.SaveHistoryEntry
	usecase.FindHistory
	usecase.ClearHistory
}
