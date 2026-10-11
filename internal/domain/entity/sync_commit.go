package entity

import "time"

type SyncCommit struct {
	When        time.Time
	Files       map[string]string
	Branch      string
	Subfolder   string
	Parent      string
	Message     string
	AuthorName  string
	AuthorEmail string
}

func (s SyncCommit) LogValue() any {
	return map[string]any{"branch": s.Branch, "subfolder": s.Subfolder, "parent": s.Parent, "files": len(s.Files)}
}
