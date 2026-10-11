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
	wails.RegisterEvent[dto.ExportProgress](dto.ExportProgressEvent)
	wails.RegisterEvent[dto.ImportFilesDropped](dto.ImportFilesDroppedEvent)
	wails.RegisterEvent[dto.ImportRequested](dto.ImportRequestedEvent)
	wails.RegisterEvent[dto.ExportRequested](dto.ExportRequestedEvent)
	wails.RegisterEvent[dto.ConsoleEntries](dto.ConsoleEntriesEvent)
	wails.RegisterEvent[dto.CookiesChanged](dto.CookiesChangedEvent)
	wails.RegisterEvent[dto.GitHostKeyUnknown](dto.GitHostKeyUnknownEvent)
}
