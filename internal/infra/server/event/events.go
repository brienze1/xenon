package event

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/publisher/dto"
)

func init() {
	wails.RegisterEvent[dto.WorkspaceTreeChanged](dto.WorkspaceTreeChangedEvent)
	wails.RegisterEvent[dto.RequestChangedOnDisk](dto.RequestChangedOnDiskEvent)
	wails.RegisterEvent[dto.WorkspaceFolderMissing](dto.WorkspaceFolderMissingEvent)
	wails.RegisterEvent[dto.VariablesChanged](dto.VariablesChangedEvent)
	wails.RegisterEvent[dto.VaultStatusChanged](dto.VaultStatusChangedEvent)
	wails.RegisterEvent[dto.ImportProgress](dto.ImportProgressEvent)
}
