package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/linkmode"
)

type FolderInfo struct {
	Path             string
	EntryCount       int
	Exists           bool
	IsDirectory      bool
	HasWorkspaceFile bool
}

func (f FolderInfo) IsEmptyOrAbsent() bool {
	return !f.Exists || (f.IsDirectory && f.EntryCount == 0)
}

func (f FolderInfo) IsWorkspace() bool {
	return f.IsDirectory && f.HasWorkspaceFile
}

func (f FolderInfo) LinkModeFor(requested enums.LinkMode) enums.LinkMode {
	if !f.IsWorkspace() {
		return linkmode.PlaceMine
	}
	return requested
}
