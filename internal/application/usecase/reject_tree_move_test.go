package usecase

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func moveTree() entity.Tree {
	ping := entity.TreeNode{ID: "r1", Kind: itemkind.Request, Name: "Ping", FileName: "Hand-Named.yaml"}
	folder := entity.TreeNode{ID: "f1", Kind: itemkind.Folder, Name: "Inner", FileName: "inner", Children: []entity.TreeNode{ping}}
	billing := entity.TreeNode{ID: "c1", Kind: itemkind.Collection, Name: "Billing", FileName: "billing", Children: []entity.TreeNode{folder}}
	return entity.NewTree([]entity.TreeNode{billing}, []string{"billing"})
}

func TestResolveTopLevelTargetTakesOnlyCollections(t *testing.T) {
	tree := moveTree()
	collection, _ := tree.Item("c1")
	folder, _ := tree.Item("f1")
	if target, err := ResolveTopLevelTarget(tree, *collection); err != nil || target.ID != tree.Root.ID {
		t.Fatalf("collection = %+v, %v", target, err)
	}
	if _, err := ResolveTopLevelTarget(tree, *folder); codeOf(err) != "TRE-38422" {
		t.Fatalf("folder = %v", err)
	}
}

func TestRejectTreeMoves(t *testing.T) {
	tree := moveTree()
	collection, _ := tree.Item("c1")
	folder, _ := tree.Item("f1")
	request, _ := tree.Item("r1")
	if err := RejectNestedCollection(*collection); codeOf(err) != "TRE-39422" {
		t.Fatalf("nested collection = %v", err)
	}
	if err := RejectNestedCollection(*folder); err != nil {
		t.Fatalf("folder = %v", err)
	}
	if err := RejectMoveIntoItself(*folder, *folder); codeOf(err) != "TRE-02422" {
		t.Fatalf("into itself = %v", err)
	}
	if err := RejectMoveIntoItself(*request, *collection); err != nil {
		t.Fatalf("request into collection = %v", err)
	}
}
