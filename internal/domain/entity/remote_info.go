package entity

const fallbackBranch = "main"

type RemoteInfo struct {
	DefaultBranch string
	Branches      []string
}

func (r RemoteInfo) CloneBranch() string {
	if r.DefaultBranch == "" {
		return fallbackBranch
	}
	return r.DefaultBranch
}
