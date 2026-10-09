package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type WorkspacePersistence interface {
	usecase.LoadWorkspace
	usecase.SaveWorkspaceFile
	usecase.ValidateWorkspaceFolder
}
