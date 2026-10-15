package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type ImportSourcePersistence interface {
	usecase.ReadImportSource
	usecase.ReadImportFolderFiles
	usecase.SaveImportSource
	usecase.FindImportSource
	usecase.DiscardImportSource
}
