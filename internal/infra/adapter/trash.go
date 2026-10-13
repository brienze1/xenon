package adapter

import (
	"os"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type osTrash struct{}

func NewTrash() adapter.Trash {
	return &osTrash{}
}

func (o *osTrash) MoveToTrash(path string) error {
	if _, err := os.Lstat(path); err != nil {
		return err
	}
	return moveToTrash(path)
}
