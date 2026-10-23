package entity

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func TestFlattenRunRowsWalksTheSidebarDepthFirst(t *testing.T) {
	target := TreeNode{ID: "shop", Kind: itemkind.Collection, Children: []TreeNode{
		{ID: "login", Kind: itemkind.Request, Name: "Login", Method: "POST", Examples: []TreeExample{{}}},
		{ID: "orders", Kind: itemkind.Folder, Name: "Orders", Children: []TreeNode{
			{ID: "create", Kind: itemkind.Request, Name: "Create", Method: "POST"},
			{ID: "nested", Kind: itemkind.Folder, Name: "Nested", Children: []TreeNode{{ID: "deep", Kind: itemkind.Request, Name: "Deep", Method: "GET"}}},
			{ID: "error:broken.yaml", Kind: itemkind.Error, Name: "broken.yaml"},
		}},
		{ID: "ignored", Kind: itemkind.Request, Name: "Ignored", Ignored: true},
		{ID: "logout", Kind: itemkind.Request, Name: "Logout", Method: "POST"},
	}}
	rows := FlattenRunRows(target)
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.RequestID)
	}
	if !reflect.DeepEqual(ids, []string{"login", "create", "deep", "logout"}) {
		t.Fatalf("ids = %v", ids)
	}
	if rows[0].Number != 1 || rows[3].Number != 4 || !rows[0].Selected || rows[0].Method != "POST" {
		t.Errorf("rows = %+v", rows)
	}
	if len(rows[0].FolderPath) != 0 || !reflect.DeepEqual(rows[1].FolderPath, []string{"Orders"}) || !reflect.DeepEqual(rows[2].FolderPath, []string{"Orders", "Nested"}) {
		t.Errorf("folder paths = %v, %v, %v", rows[0].FolderPath, rows[1].FolderPath, rows[2].FolderPath)
	}
}

func TestFlattenRunRowsOfAFolderIsRelativeToTheFolder(t *testing.T) {
	folder := TreeNode{ID: "orders", Kind: itemkind.Folder, Name: "Orders", Children: []TreeNode{
		{ID: "create", Kind: itemkind.Request, Name: "Create"},
		{ID: "nested", Kind: itemkind.Folder, Name: "Nested", Children: []TreeNode{{ID: "deep", Kind: itemkind.Request, Name: "Deep"}}},
	}}
	rows := FlattenRunRows(folder)
	if len(rows) != 2 || len(rows[0].FolderPath) != 0 || !reflect.DeepEqual(rows[1].FolderPath, []string{"Nested"}) {
		t.Errorf("rows = %+v", rows)
	}
	if found, ok := RunPlanRowOf(rows, "deep"); !ok || found.Name != "Deep" {
		t.Errorf("RunPlanRowOf = %+v, %v", found, ok)
	}
	if _, ok := RunPlanRowOf(rows, "missing"); ok {
		t.Error("a missing request is not found")
	}
	if len(FlattenRunRows(TreeNode{})) != 0 {
		t.Error("an empty target has no rows")
	}
}
