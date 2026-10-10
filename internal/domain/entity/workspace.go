package entity

import "strings"

type Workspace struct {
	ID          string
	Name        string
	Description string
	Order       []string
	Variables   []VariableRow
}

func NewWorkspace(id, name string) Workspace {
	return Workspace{ID: id, Name: strings.TrimSpace(name), Order: []string{}, Variables: []VariableRow{}}
}

func (w Workspace) HasValidName() bool {
	return isValidItemName(w.Name)
}

func (w Workspace) ArrangedCollections(present []string) []string {
	return arrangedNames(w.Order, present)
}

func (w Workspace) Renamed(name string) Workspace {
	w.Name = strings.TrimSpace(name)
	return w
}
