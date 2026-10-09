package entity

import "strings"

type Workspace struct {
	ID          string
	Name        string
	Description string
	Order       []string
}

func NewWorkspace(id, name string) Workspace {
	return Workspace{ID: id, Name: strings.TrimSpace(name), Order: []string{}}
}

func (w Workspace) HasValidName() bool {
	return isValidItemName(w.Name)
}
