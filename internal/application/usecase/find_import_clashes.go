package usecase

import (
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
		if match, found := entity.NamedTreeNode(existing, node.Name()); found {
			clashes = append(clashes, entity.Clash{ItemKey: itemKey, Kind: clashKinds[node.Kind()], ExistingID: match.ID, ExistingName: match.Name, Name: node.Name()})
		}
	}
	return clashes
}
