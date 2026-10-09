package adapter

import (
	"github.com/brienze1/xenon/internal/integration/adapter"
)

type osTrash struct{}

func NewTrash() adapter.Trash {
	return &osTrash{}
}

func (o *osTrash) MoveToTrash(path string) error {
	return moveToTrash(path)
}
