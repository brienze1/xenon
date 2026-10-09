package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

type Collection struct {
	Auth        Auth
	ID          string
	Kind        enums.ItemKind
	Name        string
	Description string
	Order       []string
	Headers     []KeyValue
}

func NewCollection(id, name string) Collection {
	return newContainer(id, name, itemkind.Collection)
}

func NewFolder(id, name string) Collection {
	return newContainer(id, name, itemkind.Folder)
}

func newContainer(id, name string, kind enums.ItemKind) Collection {
	return Collection{
		ID:      id,
		Kind:    kind,
		Name:    strings.TrimSpace(name),
		Order:   []string{},
		Auth:    DefaultAuth(),
		Headers: []KeyValue{},
	}
}

func (c Collection) IsFolder() bool {
	return c.Kind == itemkind.Folder
}

func (c Collection) HasValidName() bool {
	return isValidItemName(c.Name)
}
