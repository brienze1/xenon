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

func runPlanRowIds(rows []RunPlanRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.RequestID)
	}
	return ids
}

func TestReconcileRunRowsDropsRemovedRequestsAndKeepsTheRememberedOrder(t *testing.T) {
	sidebar := []RunPlanRow{{RunRow: RunRow{RequestID: "login"}, Name: "Login"}, {RunRow: RunRow{RequestID: "create"}}, {RunRow: RunRow{RequestID: "get"}}}
	remembered := []RunRow{{RequestID: "logout", Selected: true}, {RequestID: "get"}, {RequestID: "login", Selected: true}, {RequestID: "create", Selected: true}}
	rows := ReconcileRunRows(remembered, sidebar)
	if !reflect.DeepEqual(runPlanRowIds(rows), []string{"get", "login", "create"}) {
		t.Fatalf("ids = %v", runPlanRowIds(rows))
	}
	if rows[0].Selected || !rows[1].Selected || rows[1].Name != "Login" || rows[2].Number != 3 {
		t.Errorf("rows = %+v", rows)
	}
}

func TestReconcileRunRowsInsertsNewRequestsSelectedAfterTheirPrecedingSibling(t *testing.T) {
	sidebar := []RunPlanRow{{RunRow: RunRow{RequestID: "login"}}, {RunRow: RunRow{RequestID: "create"}}, {RunRow: RunRow{RequestID: "refund"}}, {RunRow: RunRow{RequestID: "audit"}}, {RunRow: RunRow{RequestID: "get"}}}
	remembered := []RunRow{{RequestID: "get"}, {RequestID: "login", Selected: true}, {RequestID: "create", Selected: true}}
	rows := ReconcileRunRows(remembered, sidebar)
	if !reflect.DeepEqual(runPlanRowIds(rows), []string{"get", "login", "create", "refund", "audit"}) {
		t.Fatalf("ids = %v", runPlanRowIds(rows))
	}
	if !rows[3].Selected || !rows[4].Selected || rows[4].Number != 5 {
		t.Errorf("new rows = %+v", rows[3:])
	}
}

func TestReconcileRunRowsInsertsANewRequestFirstWithoutAPresentPredecessor(t *testing.T) {
	sidebar := []RunPlanRow{{RunRow: RunRow{RequestID: "first"}}, {RunRow: RunRow{RequestID: "login"}}}
	rows := ReconcileRunRows([]RunRow{{RequestID: "login", Selected: true}}, sidebar)
	if !reflect.DeepEqual(runPlanRowIds(rows), []string{"first", "login"}) {
		t.Fatalf("ids = %v", runPlanRowIds(rows))
	}
	if all := ReconcileRunRows(nil, sidebar); !reflect.DeepEqual(runPlanRowIds(all), []string{"first", "login"}) || !all[1].Selected {
		t.Errorf("nothing remembered lists the sidebar, selected: %+v", all)
	}
}

func TestReconcileRunRowsFallsBackToTheDefaultForAStaleAlias(t *testing.T) {
	sidebar := []RunPlanRow{
		RunPlanRow{RunRow: RunRow{RequestID: "login"}}.WithAliases([]RequestAlias{{ID: "mx", Name: "mx-test-user", Version: Request{URL: "u"}}}),
		{RunRow: RunRow{RequestID: "get"}, Aliases: []RequestAlias{}},
	}
	rows := ReconcileRunRows([]RunRow{{RequestID: "login", AliasID: "mx", Selected: true}, {RequestID: "get", AliasID: "gone", Selected: true}}, sidebar)
	if rows[0].AliasID != "mx" || rows[1].AliasID != "" {
		t.Errorf("aliases = %q, %q", rows[0].AliasID, rows[1].AliasID)
	}
	if rows[0].Aliases[0].Version.URL != "" || rows[0].Aliases[0].Name != "mx-test-user" {
		t.Errorf("plan aliases keep only id and name: %+v", rows[0].Aliases)
	}
}

func TestFindNextRunRowMatchesTheIdThenTheFirstName(t *testing.T) {
	rows := []RunPlanRow{{RunRow: RunRow{RequestID: "a"}, Name: "Same"}, {RunRow: RunRow{RequestID: "b"}, Name: "a"}, {RunRow: RunRow{RequestID: "c"}, Name: "Same"}}
	if index, found := FindNextRunRow(rows, "a"); !found || index != 0 {
		t.Errorf("an id wins over a name: %d, %v", index, found)
	}
	if index, found := FindNextRunRow(rows, "Same"); !found || index != 0 {
		t.Errorf("the first row with the name: %d, %v", index, found)
	}
	if _, found := FindNextRunRow(rows, "Nope"); found {
		t.Error("an unknown target is not found")
	}
}

func TestNextRunPositionFollowsSetNextRequest(t *testing.T) {
	rows := []RunPlanRow{{RunRow: RunRow{RequestID: "a"}, Name: "A"}, {RunRow: RunRow{RequestID: "b"}, Name: "B"}, {RunRow: RunRow{RequestID: "c"}, Name: "C"}}
	cases := []struct {
		next     *NextRequest
		name     string
		position int
		known    bool
	}{
		{name: "no call", position: 2, known: true},
		{name: "by name", next: &NextRequest{Target: "A"}, position: 0, known: true},
		{name: "by id", next: &NextRequest{Target: "c"}, position: 2, known: true},
		{name: "null", next: &NextRequest{Stop: true}, position: 3, known: true},
		{name: "unknown", next: &NextRequest{Target: "Nope"}, position: 3, known: false},
	}
	for _, test := range cases {
		if position, known := NextRunPosition(rows, 1, test.next); position != test.position || known != test.known {
			t.Errorf("%s: position %d, known %v", test.name, position, known)
		}
	}
	if warning := UnknownNextRequestWarning("Nope"); warning != `Couldn't find request "Nope" in this run; the iteration ends.` {
		t.Errorf("warning = %q", warning)
	}
}
