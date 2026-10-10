package entity

import (
	"sort"
	"strconv"
)

type EnvironmentList struct {
	ActiveID     string
	Environments []Environment
}

// NewEnvironmentList sorts by name and drops an active id that matches no readable environment.
func NewEnvironmentList(environments []Environment, activeId string) EnvironmentList {
	sorted := append([]Environment{}, environments...)
	sort.SliceStable(sorted, func(left, right int) bool {
		return sorted[left].sortKey() < sorted[right].sortKey()
	})
	list := EnvironmentList{Environments: sorted}
	if list.Contains(activeId) {
		list.ActiveID = activeId
	}
	return list
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

func (e EnvironmentList) FileNames(exceptId string) []string {
	names := make([]string, 0, len(e.Environments))
	for _, environment := range e.Environments {
		if environment.IsMalformed() || environment.ID != exceptId {
			names = append(names, environment.FileName)
		}
	}
	return names
}
