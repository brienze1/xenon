package event

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/publisher/dto"
)

func init() {
	wails.RegisterEvent[dto.WorkspaceTreeChanged](dto.WorkspaceTreeChangedEvent)
	wails.RegisterEvent[dto.RequestChangedOnDisk](dto.RequestChangedOnDiskEvent)
	wails.RegisterEvent[dto.WorkspaceFolderMissing](dto.WorkspaceFolderMissingEvent)
}
