package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type IterationDataPersistence interface {
	usecase.FindIterationData
}
