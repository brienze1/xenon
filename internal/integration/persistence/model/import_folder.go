package model

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportFolder struct {
	Name    string
	Entries []string
}

func (i *ImportFolder) ToEntity() entity.ImportProbe {
	return entity.ImportProbe{Name: i.Name, IsFolder: true, Entries: i.Entries}
}

func NewImportFolder(name string, entries []string) ImportFolder {
	return ImportFolder{Name: name, Entries: entries}
}
