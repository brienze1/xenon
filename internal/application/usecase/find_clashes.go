package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/clashkind"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

var clashKinds = map[enums.ItemKind]enums.ClashKind{
	itemkind.Collection: clashkind.Collection,
	itemkind.Folder:     clashkind.Folder,
	itemkind.Request:    clashkind.Request,
}

func FindImportClashes(itemKey string, nodes []entity.ImportNode, existing []entity.TreeNode) []entity.Clash {
	clashes := make([]entity.Clash, 0)
	for _, node := range nodes {
		if match, found := sameNamed(node.Name(), existing); found {
			clashes = append(clashes, entity.Clash{ItemKey: itemKey, Kind: clashKinds[node.Kind()], ExistingID: match.ID, ExistingName: match.Name, Name: node.Name()})
		}
	}
	return clashes
}

func FindVariableClashes(itemKey string, imported []entity.VariableRow, collection entity.Collection) []entity.Clash {
	clashes := make([]entity.Clash, 0)
	for _, row := range imported {
		if _, found := variableIndex(collection.Variables, row.Key); found {
			clashes = append(clashes, entity.Clash{ItemKey: itemKey, Kind: clashkind.Variable, ExistingID: collection.ID, ExistingName: row.Key, Name: row.Key})
		}
	}
	return clashes
}

func sameNamed(name string, existing []entity.TreeNode) (entity.TreeNode, bool) {
	for _, node := range existing {
		if node.IsItem() && strings.EqualFold(strings.TrimSpace(node.Name), strings.TrimSpace(name)) {
			return node, true
		}
	}
	return entity.TreeNode{}, false
}

func variableIndex(rows []entity.VariableRow, key string) (int, bool) {
	for index, row := range rows {
		if row.Key == key {
			return index, true
		}
	}
	return -1, false
}
