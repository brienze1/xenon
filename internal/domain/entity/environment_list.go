package entity

import (
	"slices"
	"sort"
	"strconv"
)

type EnvironmentList struct {
	ActiveID     string
	Environments []Environment
}

// NewEnvironmentList arranges by order, then by name, broken files last, and drops an active id that matches no readable environment.
func NewEnvironmentList(environments []Environment, order []string, activeId string) EnvironmentList {
	list := EnvironmentList{Environments: arrangedEnvironments(environments, order)}
	if list.Contains(activeId) {
		list.ActiveID = activeId
	}
	return list
}

func arrangedEnvironments(environments []Environment, order []string) []Environment {
	readable := make([]Environment, 0, len(environments))
	malformed := make([]Environment, 0)
	for _, environment := range environments {
		if environment.IsMalformed() {
			malformed = append(malformed, environment)
		} else {
			readable = append(readable, environment)
		}
	}
	sort.SliceStable(readable, func(left, right int) bool {
		return readable[left].sortKey() < readable[right].sortKey()
	})
	sort.SliceStable(malformed, func(left, right int) bool {
		return malformed[left].FileName < malformed[right].FileName
	})
	arranged := make([]Environment, 0, len(environments))
	for _, id := range order {
		index := slices.IndexFunc(readable, func(environment Environment) bool { return environment.HasID(id) })
		if index >= 0 {
			arranged = append(arranged, readable[index])
			readable = slices.Delete(readable, index, index+1)
		}
	}
	return append(append(arranged, readable...), malformed...)
}

func (e EnvironmentList) IDs() []string {
	ids := make([]string, 0, len(e.Environments))
	for _, environment := range e.Environments {
		if !environment.IsMalformed() {
			ids = append(ids, environment.ID)
		}
	}
	return ids
}

func (e EnvironmentList) IsOrderedAs(order []string) bool {
	return slices.Equal(e.IDs(), order)
}

// OrderMoving places id at index among the other readable environments; an index past the end places it last.
func (e EnvironmentList) OrderMoving(id string, index int) []string {
	order := e.OrderRemoving(id)
	return slices.Insert(order, min(max(index, 0), len(order)), id)
}

func (e EnvironmentList) OrderAppending(id string) []string {
	order := e.IDs()
	if slices.Contains(order, id) {
		return order
	}
	return append(order, id)
}

// OrderReplacing swaps previous for id in place, or appends id when previous is not listed.
func (e EnvironmentList) OrderReplacing(previous, id string) []string {
	order := e.IDs()
	index := slices.Index(order, previous)
	if index < 0 {
		return e.OrderAppending(id)
	}
	order[index] = id
	return order
}

func (e EnvironmentList) OrderRemoving(id string) []string {
	return slices.DeleteFunc(e.IDs(), func(listed string) bool { return listed == id })
}

func (e EnvironmentList) Contains(id string) bool {
	for _, environment := range e.Environments {
		if environment.HasID(id) {
			return true
		}
	}
	return false
}

func (e EnvironmentList) LogValue() any {
	return map[string]any{"activeId": e.ActiveID, "environments": len(e.Environments)}
}

func (e EnvironmentList) IsNameTaken(name, exceptId string) bool {
	for _, environment := range e.Environments {
		if !environment.IsMalformed() && environment.ID != exceptId && environment.HasName(name) {
			return true
		}
	}
	return false
}

func (e EnvironmentList) CopyName(name string) string {
	candidate := name + copySuffix
	for suffix := 2; e.IsNameTaken(candidate, ""); suffix++ {
		candidate = name + copySuffix + " " + strconv.Itoa(suffix)
	}
	return candidate
}

func (e EnvironmentList) KeepsName(environment Environment) bool {
	for _, existing := range e.Environments {
		if existing.HasID(environment.ID) && existing.HasName(environment.Name) {
			return true
		}
	}
	return false
}

func (e EnvironmentList) IsNameAvailableFor(environment Environment) bool {
	return !e.IsNameTaken(environment.Name, environment.ID)
}

func (e EnvironmentList) SiblingFileNames(environment Environment) []string {
	names := make([]string, 0, len(e.Environments))
	for _, existing := range e.Environments {
		if !existing.HasID(environment.ID) {
			names = append(names, existing.FileName)
		}
	}
	return names
}
