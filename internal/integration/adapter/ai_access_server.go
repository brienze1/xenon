package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type AIAccessServer interface {
	usecase.ServeAIAccess
}
