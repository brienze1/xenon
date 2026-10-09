package entity

import (
	"path/filepath"
	"strings"
	"time"
)

type RegisteredWorkspace struct {
	AddedAt          time.Time
	LocalID          string
	Path             string
	DisplayNameCache string
	Missing          bool
}

func NewRegisteredWorkspace(localId, path, name string, addedAt time.Time) RegisteredWorkspace {
	return RegisteredWorkspace{LocalID: localId, Path: filepath.Clean(path), DisplayNameCache: name, AddedAt: addedAt}
}

func (r RegisteredWorkspace) Overlaps(path string) bool {
	return isSameOrInside(path, r.Path) || isSameOrInside(r.Path, path)
}

func isSameOrInside(child, parent string) bool {
	relative, err := filepath.Rel(filepath.Clean(parent), filepath.Clean(child))
	if err != nil {
		return false
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
