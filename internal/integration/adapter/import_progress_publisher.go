package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type ImportProgressPublisher interface {
	usecase.PublishImportProgress
}
