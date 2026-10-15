package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type UpdateInstallPersistence interface {
	usecase.InstallUpdate
	usecase.CleanUpdateLeftovers
}
