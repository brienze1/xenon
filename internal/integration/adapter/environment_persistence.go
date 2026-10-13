package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type EnvironmentPersistence interface {
	usecase.ListEnvironments
	usecase.FindEnvironment
	usecase.SaveEnvironment
	usecase.DeleteEnvironment
}
