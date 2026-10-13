package importitemkind

import "github.com/brienze1/xenon/internal/domain/enums"

const (
	Collection  enums.ImportItemKind = "collection"
	Folder      enums.ImportItemKind = "folder"
	Workspace   enums.ImportItemKind = "workspace"
	Environment enums.ImportItemKind = "environment"
	Globals     enums.ImportItemKind = "globals"
	Request     enums.ImportItemKind = "request"
	Unknown     enums.ImportItemKind = "unknown"
)
