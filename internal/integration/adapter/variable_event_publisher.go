package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type VariableEventPublisher interface {
	usecase.PublishVariableEvent
}
