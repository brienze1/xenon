package entity

import "sort"

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
