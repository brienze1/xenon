package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type RequestPersistence interface {
	usecase.FindRequest
	usecase.SaveRequest
}
