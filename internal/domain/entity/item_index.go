package entity

import (
	"path"
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

const yamlExtension = ".yaml"

type ItemIndex map[string]ItemRef

func (i ItemIndex) RefFor(filePath string) ItemRef {
	if ref, indexed := i[filePath]; indexed {
		return ref
	}
	if workspace, found := i[workspaceFileName]; found && isWorkspaceLevel(filePath) {
		return workspace
	}
	return itemRefFromPath(filePath)
}

func (i ItemIndex) ChangeItems(changes []PathChange) []ChangeItem {
	order := make([]string, 0, len(changes))
	items := map[string]ChangeItem{}
	for _, change := range changes {
		ref := i.RefFor(change.Path)
		existing, seen := items[ref.ID]
		if !seen {
			order = append(order, ref.ID)
			items[ref.ID] = NewChangeItem(ref, change)
			continue
		}
		items[ref.ID] = existing.Combined(change)
	}
	mapped := make([]ChangeItem, 0, len(order))
	for _, id := range order {
		mapped = append(mapped, items[id])
	}
	return mapped
}

func itemRefFromPath(filePath string) ItemRef {
	kind := PathPlan{Path: filePath}.StructuralKind()
	name := strings.TrimSuffix(path.Base(filePath), yamlExtension)
	if kind == itemkind.Collection || kind == itemkind.Folder {
		name = path.Base(path.Dir(filePath))
	}
	if isWorkspaceLevel(filePath) {
		kind = itemkind.Workspace
	}
	return ItemRef{ID: filePath, Name: name, Kind: requestWhenUnknown(kind)}
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
