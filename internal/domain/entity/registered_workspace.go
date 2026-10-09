package entity

import (
	"path/filepath"
	"strings"
	"time"
)

const (
	temporaryFileSuffix = ".xenon-tmp"
	workspaceFileName   = "workspace.yaml"
	collectionsFolder   = "collections"
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

func (r RegisteredWorkspace) Watches(path string) bool {
	if filepath.Clean(path) == r.Path {
		return true
	}
	if strings.HasSuffix(path, temporaryFileSuffix) || !isSameOrInside(path, r.Path) {
		return false
	}
	relative, _ := filepath.Rel(r.Path, path)
	segments := strings.Split(filepath.ToSlash(relative), "/")
	return relative == workspaceFileName || segments[0] == collectionsFolder
}
