package event

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/publisher/dto"
)

const (
	WorkspaceTreeChanged   = "workspace:tree-changed"
	RequestChangedOnDisk   = "request:changed-on-disk"
	WorkspaceFolderMissing = "workspace:folder-missing"
)

func init() {
	wails.RegisterEvent[dto.WorkspaceTreeChanged](WorkspaceTreeChanged)
	wails.RegisterEvent[dto.RequestChangedOnDisk](RequestChangedOnDisk)
	wails.RegisterEvent[dto.WorkspaceFolderMissing](WorkspaceFolderMissing)
}
