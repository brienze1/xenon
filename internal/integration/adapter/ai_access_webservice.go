package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type AIAccessWebservice interface {
	usecase.ServeAIAccess
	usecase.FindAIAccessStatus
}
