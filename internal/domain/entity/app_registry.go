package entity

import "path/filepath"

const registryVersion = 1

type AppRegistry struct {
	ActiveLocalID string
	Workspaces    []RegisteredWorkspace
	Version       int
}

func NewAppRegistry() AppRegistry {
	return AppRegistry{Version: registryVersion, Workspaces: []RegisteredWorkspace{}}
}

func (a *AppRegistry) Overlaps(path string) bool {
	for _, workspace := range a.Workspaces {
		if workspace.Overlaps(path) {
			return true
		}
	}
	return false
}

func (a *AppRegistry) Register(workspace RegisteredWorkspace) {
	a.Workspaces = append(a.Workspaces, workspace)
	a.ActiveLocalID = workspace.LocalID
}

func (a *AppRegistry) Find(localId string) (*RegisteredWorkspace, bool) {
	for index := range a.Workspaces {
		if a.Workspaces[index].LocalID == localId {
			return &a.Workspaces[index], true
		}
	}
	return nil, false
}

func (a *AppRegistry) Activate(localId, displayName string) (*RegisteredWorkspace, bool) {
	workspace, found := a.Find(localId)
	if !found {
		return nil, false
	}
	workspace.DisplayNameCache = displayName
	workspace.Missing = false
	a.ActiveLocalID = localId
	return workspace, true
}

func (a *AppRegistry) OverlapsOther(path, localId string) bool {
	for _, workspace := range a.Workspaces {
		if workspace.LocalID != localId && workspace.Overlaps(path) {
			return true
		}
	}
	return false
}

func (a *AppRegistry) Rename(localId, displayName string) (*RegisteredWorkspace, bool) {
	workspace, found := a.Find(localId)
	if !found {
		return nil, false
	}
	workspace.DisplayNameCache = displayName
	return workspace, true
}

func (a *AppRegistry) Relocate(localId, path, displayName string) (*RegisteredWorkspace, bool) {
	workspace, found := a.Find(localId)
	if !found {
		return nil, false
	}
	workspace.Path = filepath.Clean(path)
	workspace.DisplayNameCache = displayName
	workspace.Missing = false
	return workspace, true
}

func (a *AppRegistry) Remove(localId string) bool {
	for index, workspace := range a.Workspaces {
		if workspace.LocalID != localId {
			continue
		}
		a.Workspaces = append(a.Workspaces[:index], a.Workspaces[index+1:]...)
		if a.ActiveLocalID == localId {
			a.ActiveLocalID = a.neighbourOf(index)
		}
		return true
	}
	return false
}

func (a *AppRegistry) neighbourOf(index int) string {
	if index < len(a.Workspaces) {
		return a.Workspaces[index].LocalID
	}
	if index > 0 {
		return a.Workspaces[index-1].LocalID
	}
	return ""
}

func (a *AppRegistry) IsActive(localId string) bool {
	return a.ActiveLocalID == localId
}
