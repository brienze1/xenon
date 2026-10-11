package entity

type GitAccess struct {
	Repository      string
	Secret          string
	TrustedHostKeys []HostKey
	Account         GitAccount
}

func NewGitAccess(repository string, account GitAccount, secret string) GitAccess {
	return GitAccess{Repository: repository, Account: account, Secret: secret}
}

func (g GitAccess) WithTrustedHostKeys(keys []HostKey) GitAccess {
	g.TrustedHostKeys = keys
	return g
}

func (g GitAccess) LogValue() any {
	return map[string]any{"repository": g.Repository, "accountId": g.Account.ID, "hasSecret": g.Secret != "", "trustedHostKeys": len(g.TrustedHostKeys)}
}
