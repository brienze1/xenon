package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type AISessionContextPersistence interface {
	usecase.FindAISessionContext
	usecase.SaveAISessionContext
	usecase.DeleteAISessionContext
}
