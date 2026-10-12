package entity

import "path/filepath"

type LocalSyncSettings struct {
	AccountID  string
	ClonePath  string
	Repository string
	Branch     string
	Subfolder  string
	AutoPull   bool
	AutoPush   bool
}

func NewLocalSyncSettings(shared SyncSettings, accountId, clonePath string) LocalSyncSettings {
	return LocalSyncSettings{
		AccountID:  accountId,
		ClonePath:  filepath.Clean(clonePath),
		Repository: shared.Repository,
		Branch:     shared.Branch,
		Subfolder:  shared.Subfolder,
	}
}

func (l LocalSyncSettings) UsesAccount(accountId string) bool {
	return l.AccountID != "" && l.AccountID == accountId
}

func (l LocalSyncSettings) WithoutAccount() LocalSyncSettings {
	l.AccountID = ""
	return l
}

func (l LocalSyncSettings) WithOptions(accountId string, autoPull, autoPush bool) LocalSyncSettings {
	l.AccountID = accountId
	l.AutoPull = autoPull
	l.AutoPush = autoPush
	return l
}

func (l LocalSyncSettings) IsLinked() bool {
	return l.ClonePath != "" && l.Branch != "" && l.Repository != ""
}

func (l LocalSyncSettings) HasAccount() bool {
	return l.AccountID != ""
}

func (l LocalSyncSettings) Shared() SyncSettings {
	return SyncSettings{Repository: l.Repository, Branch: l.Branch, Subfolder: l.Subfolder}
}

func (l LocalSyncSettings) WorkspacePath() string {
	return filepath.Join(l.ClonePath, filepath.FromSlash(l.Subfolder))
}

func (l LocalSyncSettings) PullsWhenOpened() bool {
	return l.IsLinked() && l.AutoPull
}

func (l LocalSyncSettings) PushesAfterSave() bool {
	return l.IsLinked() && l.AutoPush
}

func (l LocalSyncSettings) Encloses(path string) bool {
	return isSameOrInside(path, l.ClonePath)
}
