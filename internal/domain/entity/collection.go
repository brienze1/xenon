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
	Variables   []VariableRow
}

func NewCollection(id, name string) Collection {
	return newContainer(id, name, itemkind.Collection)
}

func NewFolder(id, name string) Collection {
	return newContainer(id, name, itemkind.Folder)
}

func newContainer(id, name string, kind enums.ItemKind) Collection {
	return Collection{
		ID:        id,
		Kind:      kind,
		Name:      strings.TrimSpace(name),
		Order:     []string{},
		Auth:      DefaultAuth(),
		Headers:   []KeyValue{},
		Variables: []VariableRow{},
	}
}

func (c Collection) IsFolder() bool {
	return c.Kind == itemkind.Folder
}

func (c Collection) HasValidName() bool {
	return isValidItemName(c.Name)
}

func (c Collection) WithSettings(settings Collection) Collection {
	c.Description = settings.Description
	c.Auth = settings.Auth
	c.Headers = append([]KeyValue{}, settings.Headers...)
	return c
}

func (c Collection) WithOrder(order []string) Collection {
	c.Order = append([]string{}, order...)
	return c
}

func (c Collection) WithVariables(rows []VariableRow) Collection {
	c.Variables = append([]VariableRow{}, rows...)
	return c
}

func (c Collection) OwnVariables() []VariableRow {
	if c.IsFolder() {
		return []VariableRow{}
	}
	return c.Variables
}

func (c Collection) ScopeRef() ScopeRef {
	return NewCollectionScopeRef(c.ID)
}

func (c Collection) MapText(transform func(string) string) Collection {
	c.Auth = c.Auth.MapText(transform)
	c.Headers = mapKeyValues(c.Headers, transform)
	return c
}

func (c Collection) ArrangedChildren(present []string) []string {
	return arrangedNames(c.Order, present)
}

func (c Collection) LogValue() any {
	return map[string]any{"id": c.ID, "kind": c.Kind, "name": c.Name, "description": c.Description, "headers": KeyValues(c.Headers).LogValue()}
}
