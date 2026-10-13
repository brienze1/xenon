package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type RegistryPersistence interface {
	usecase.FindWorkspaceRegistry
	usecase.SaveWorkspaceRegistry
	usecase.FindWindowState
	usecase.SaveWindowState
	usecase.FindAppSettings
	usecase.SaveAppSettings
}
