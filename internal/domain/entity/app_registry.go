package entity

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
