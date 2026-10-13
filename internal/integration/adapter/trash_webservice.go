package adapter

import "github.com/brienze1/xenon/internal/application/usecase"

type TrashWebservice interface {
	usecase.MoveToTrash
}
