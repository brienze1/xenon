package entity

import "strings"

const postmanGlobalsKey = "globals"

type Workspace struct {
	Postman     SourceFields
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

func (w Workspace) WithVariables(rows []VariableRow) Workspace {
	w.Variables = append([]VariableRow{}, rows...)
	return w
}

func (w Workspace) WithPostmanGlobals(fields SourceFields) Workspace {
	if len(fields) == 0 {
		return w
	}
	postman := SourceFields{postmanGlobalsKey: map[string]any(fields)}
	for key, value := range w.Postman {
		if key != postmanGlobalsKey {
			postman[key] = value
		}
	}
	w.Postman = postman
	return w
}
