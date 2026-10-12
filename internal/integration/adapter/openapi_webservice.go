package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type OpenApiWebservice interface {
	usecase.WriteApiDocument
	usecase.ReadApiDocument
}
