package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type WorkspacePersistence interface {
	usecase.LoadWorkspace
	usecase.SaveWorkspaceFile
	usecase.InspectWorkspaceFolder
	usecase.CopyWorkspaceFolder
	usecase.RemoveWorkspaceFolder
	usecase.SaveWorkspaceSync
	usecase.ClearWorkspaceFolder
}
