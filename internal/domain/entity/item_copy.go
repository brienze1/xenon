package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type ItemCopy struct {
	ID       string
	Name     string
	FileName string
	Kind     enums.ItemKind
	Children []ItemCopy
}

func (i ItemCopy) IsRequest() bool {
	return i.Kind == itemkind.Request
}

func (i ItemCopy) IsContainer() bool {
	return i.Kind == itemkind.Collection || i.Kind == itemkind.Folder
}

func (i ItemCopy) Renames() bool {
	return i.Name != ""
}
