package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type RegistryPersistence interface {
	usecase.FindWorkspaceRegistry
	usecase.SaveWorkspaceRegistry
}
