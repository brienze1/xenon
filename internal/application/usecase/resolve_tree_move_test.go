package usecase

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func moveTree() entity.Tree {
	ping := entity.TreeNode{ID: "r1", Kind: itemkind.Request, Name: "Ping", FileName: "Hand-Named.yaml"}
	status := entity.TreeNode{ID: "r2", Kind: itemkind.Request, Name: "Status", FileName: "status.yaml"}
	billing := entity.TreeNode{ID: "c1", Kind: itemkind.Collection, Name: "Billing", FileName: "billing", Children: []entity.TreeNode{ping, status}, Entries: []string{"collection.yaml", "Hand-Named.yaml", "status.yaml"}}
	users := entity.TreeNode{ID: "c2", Kind: itemkind.Collection, Name: "Users", FileName: "users", Entries: []string{"collection.yaml", "status.yaml"}}
	return entity.NewTree([]entity.TreeNode{billing, users}, []string{"billing", "users"})
}

func TestResolveTreeMoveKeepsTheFileNameUnlessTheNewParentHoldsIt(t *testing.T) {
	cases := []struct {
		id, parent, fileName string
	}{
		{id: "r1", parent: "c2", fileName: "Hand-Named.yaml"},
		{id: "r2", parent: "c2", fileName: "status-2.yaml"},
		{id: "r1", parent: "c1", fileName: "Hand-Named.yaml"},
		{id: "r2", parent: "c1", fileName: "status.yaml"},
	}
	for _, tc := range cases {
		move, err := ResolveTreeMove(moveTree(), tc.id, tc.parent, 0)
		if err != nil || move.FileName != tc.fileName {
			t.Errorf("ResolveTreeMove(%s → %s) = %+v, %v; want file %s", tc.id, tc.parent, move, err, tc.fileName)
		}
	}
}
