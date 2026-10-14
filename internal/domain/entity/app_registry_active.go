package entity

func (a *AppRegistry) Active() (RegisteredWorkspace, bool) {
	workspace, found := a.Find(a.ActiveLocalID)
	if !found {
		return RegisteredWorkspace{}, false
	}
	return *workspace, true
}
