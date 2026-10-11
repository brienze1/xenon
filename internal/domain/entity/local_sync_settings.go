package entity

type LocalSyncSettings struct {
	AccountID string
	ClonePath string
	AutoPull  bool
	AutoPush  bool
}

func (l LocalSyncSettings) UsesAccount(accountId string) bool {
	return l.AccountID != "" && l.AccountID == accountId
}

func (l LocalSyncSettings) WithoutAccount() LocalSyncSettings {
	l.AccountID = ""
	return l
}
