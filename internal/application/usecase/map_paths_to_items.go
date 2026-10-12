package usecase

import (
	"path"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const (
	yamlExtension     = ".yaml"
	workspaceFilePath = "workspace.yaml"
)

func MapPathsToItems(changes []entity.PathChange, index map[string]entity.ItemRef) []entity.ChangeItem {
	order := make([]string, 0, len(changes))
	items := map[string]entity.ChangeItem{}
	for _, change := range changes {
		ref := RefForPath(change.Path, index)
		existing, seen := items[ref.ID]
		if !seen {
			order = append(order, ref.ID)
			items[ref.ID] = entity.NewChangeItem(ref, change)
			continue
		}
		items[ref.ID] = existing.Combined(change)
	}
	mapped := make([]entity.ChangeItem, 0, len(order))
	for _, id := range order {
		mapped = append(mapped, items[id])
	}
	return mapped
}

func RefForPath(filePath string, index map[string]entity.ItemRef) entity.ItemRef {
	if ref, indexed := index[filePath]; indexed {
		return ref
	}
	if workspace, found := index[workspaceFilePath]; found && isWorkspaceLevel(filePath) {
		return workspace
	}
	return itemRefFromPath(filePath)
}

func itemRefFromPath(filePath string) entity.ItemRef {
	kind := entity.PathPlan{Path: filePath}.StructuralKind()
	name := strings.TrimSuffix(path.Base(filePath), yamlExtension)
	if kind == itemkind.Collection || kind == itemkind.Folder {
		name = path.Base(path.Dir(filePath))
	}
	if isWorkspaceLevel(filePath) {
		kind = itemkind.Workspace
	}
	return entity.ItemRef{ID: filePath, Name: name, Kind: requestWhenUnknown(kind)}
}

func isWorkspaceLevel(filePath string) bool {
	return !strings.Contains(filePath, "/")
}

func requestWhenUnknown(kind enums.ItemKind) enums.ItemKind {
	if kind == itemkind.Unknown {
		return itemkind.Request
	}
	return kind
}
