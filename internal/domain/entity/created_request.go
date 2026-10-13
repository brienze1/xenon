package entity

import "strings"

// CreatedRequest is a request a documentation edit adds, with the folder names it goes under.
type CreatedRequest struct {
	Path    string
	Folders []string
	Request Request
}

func (c CreatedRequest) Folder() string {
	return strings.Join(c.Folders, TagSeparator)
}
