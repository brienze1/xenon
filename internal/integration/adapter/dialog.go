package adapter

import "github.com/brienze1/xenon/internal/domain/entity"

type Dialog interface {
	PickFolder(title string) (string, error)
	PickFile(title string, filters []entity.FileFilter) (string, error)
	PickSaveFile(title, suggestedName string) (string, error)
	PickSources(title string) ([]string, error)
}
