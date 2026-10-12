package entity

import "time"

type SyncCommit struct {
	When        time.Time
	Files       map[string]string
	Blobs       map[string]string
	Branch      string
	Subfolder   string
	Parent      string
	Message     string
	AuthorName  string
	AuthorEmail string
}

func NewSyncCommit(linked LinkedWorkspace, parent, message string, tree map[string]string, when time.Time) SyncCommit {
	return SyncCommit{
		When:        when,
		Blobs:       tree,
		Branch:      linked.Branch(),
		Subfolder:   linked.Subfolder(),
		Parent:      parent,
		Message:     message,
		AuthorName:  linked.Account.CommitName,
		AuthorEmail: linked.Account.CommitEmail,
	}
}

func (s SyncCommit) IsEmpty() bool {
	return len(s.Files) == 0 && len(s.Blobs) == 0
}

func (s SyncCommit) LogValue() any {
	return map[string]any{"branch": s.Branch, "subfolder": s.Subfolder, "parent": s.Parent, "files": len(s.Files), "blobs": len(s.Blobs)}
}
