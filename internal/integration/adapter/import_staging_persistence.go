package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type ImportStagingPersistence interface {
	usecase.StageImportItem
	usecase.CommitImportItem
	usecase.RollbackImportItem
}
