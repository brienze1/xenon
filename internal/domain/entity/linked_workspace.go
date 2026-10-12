package entity

type LinkedWorkspace struct {
	Access   GitAccess
	Account  GitAccount
	LocalID  string
	Root     string
	Settings LocalSyncSettings
}

func NewLinkedWorkspace(localId, root string, settings LocalSyncSettings) LinkedWorkspace {
	return LinkedWorkspace{LocalID: localId, Root: root, Settings: settings}
}

func (l LinkedWorkspace) WithAccess(account GitAccount, access GitAccess) LinkedWorkspace {
	l.Account = account
	l.Access = access
	return l
}

func (l LinkedWorkspace) ClonePath() string {
	return l.Settings.ClonePath
}

func (l LinkedWorkspace) Branch() string {
	return l.Settings.Branch
}

func (l LinkedWorkspace) Subfolder() string {
	return l.Settings.Subfolder
}

func (l LinkedWorkspace) LogValue() any {
	return map[string]any{"localId": l.LocalID, "root": l.Root, "repository": l.Settings.Repository, "branch": l.Settings.Branch, "subfolder": l.Settings.Subfolder}
}
