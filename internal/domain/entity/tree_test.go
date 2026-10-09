package entity

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func sampleTree() Tree {
	request := TreeNode{ID: "r1", Kind: itemkind.Request, Name: "Create invoice", URL: "https://x/INVOICES", FileName: "create.yaml"}
	broken := TreeNode{ID: "error:x", Kind: itemkind.Error, Name: "broken.yaml", FileName: "broken.yaml"}
	ignored := TreeNode{ID: "error:notes", Kind: itemkind.Error, FileName: "notes", Ignored: true}
	folder := TreeNode{ID: "f1", Kind: itemkind.Folder, Name: "Invoices", FileName: "invoices", Children: []TreeNode{request}, Entries: []string{"folder.yaml", "create.yaml"}}
	ping := TreeNode{ID: "r2", Kind: itemkind.Request, Name: "Ping", URL: "https://x/ping", FileName: "ping.yaml"}
	collection := TreeNode{
		ID: "c1", Kind: itemkind.Collection, Name: "Billing", FileName: "billing",
		Children: []TreeNode{folder, broken, ping, ignored},
		Entries:  []string{"collection.yaml", "invoices", "broken.yaml", "ping.yaml", "notes"},
	}
	return NewTree([]TreeNode{collection}, []string{"billing"})
}

func TestTreeNodeKinds(t *testing.T) {
	cases := []struct {
		node      TreeNode
		container bool
		request   bool
		item      bool
	}{
		{node: TreeNode{Kind: itemkind.Collection}, container: true, item: true},
		{node: TreeNode{Kind: itemkind.Folder}, container: true, item: true},
		{node: TreeNode{Kind: itemkind.Request}, request: true, item: true},
		{node: TreeNode{Kind: itemkind.Error}},
	}
	for _, tc := range cases {
		if tc.node.IsContainer() != tc.container || tc.node.IsRequest() != tc.request || tc.node.IsItem() != tc.item {
			t.Errorf("%s: container %v request %v item %v", tc.node.Kind, tc.node.IsContainer(), tc.node.IsRequest(), tc.node.IsItem())
		}
	}
}

func TestTreeNodeMatchesRequestsByNameOrUrlIgnoringCase(t *testing.T) {
	request := TreeNode{Kind: itemkind.Request, Name: "List Users", URL: "https://api.test/People"}
	for query, want := range map[string]bool{"users": true, " USERS ": true, "people": true, "nope": false, "": false, "  ": false} {
		if got := request.Matches(query); got != want {
			t.Errorf("Matches(%q) = %v, want %v", query, got, want)
		}
	}
	if (TreeNode{Kind: itemkind.Folder, Name: "users"}).Matches("users") {
		t.Error("folders never match")
	}
}

func TestTreeNodeCopyNameAndRequestIds(t *testing.T) {
	tree := sampleTree()
	collection := tree.Collections()[0]
	if collection.CopyName() != "Billing Copy" {
		t.Errorf("CopyName() = %q", collection.CopyName())
	}
	if got := collection.RequestIDs(); !reflect.DeepEqual(got, []string{"r1", "r2"}) {
		t.Errorf("RequestIDs() = %v", got)
	}
	if got := (TreeNode{Kind: itemkind.Folder}).RequestIDs(); got == nil || len(got) != 0 {
		t.Errorf("empty RequestIDs() = %#v", got)
	}
}

func TestTreeNodeOrderOperations(t *testing.T) {
	collection := sampleTree().Collections()[0]
	cases := map[string]struct {
		got  []string
		want []string
	}{
		"child order": {got: collection.ChildOrder(), want: []string{"invoices", "broken.yaml", "ping.yaml"}},
		"append":      {got: collection.OrderAppending("new.yaml"), want: []string{"invoices", "broken.yaml", "ping.yaml", "new.yaml"}},
		"insert":      {got: collection.OrderInsertingAfter("invoices", "copy"), want: []string{"invoices", "copy", "broken.yaml", "ping.yaml"}},
		"insert last": {got: collection.OrderInsertingAfter("ping.yaml", "copy.yaml"), want: []string{"invoices", "broken.yaml", "ping.yaml", "copy.yaml"}},
		"rename":      {got: collection.OrderRenaming("ping.yaml", "pong.yaml"), want: []string{"invoices", "broken.yaml", "pong.yaml"}},
		"remove":      {got: collection.OrderRemoving("broken.yaml"), want: []string{"invoices", "ping.yaml"}},
		"entries":     {got: collection.EntriesExcept("ping.yaml"), want: []string{"collection.yaml", "invoices", "broken.yaml", "notes"}},
	}
	for name, tc := range cases {
		if !reflect.DeepEqual(tc.got, tc.want) {
			t.Errorf("%s = %v, want %v", name, tc.got, tc.want)
		}
	}
}

func TestTreeLookups(t *testing.T) {
	tree := sampleTree()
	if item, found := tree.Item("r1"); !found || item.Name != "Create invoice" {
		t.Errorf("Item(r1) = %v %v", item, found)
	}
	if _, found := tree.Item("error:x"); found {
		t.Error("error nodes are not items")
	}
	if _, found := tree.Item("missing"); found {
		t.Error("unknown ids are not found")
	}
	if parent, found := tree.Parent("r1"); !found || parent.ID != "f1" {
		t.Errorf("Parent(r1) = %v", parent)
	}
	if parent, found := tree.Parent("c1"); !found || parent.Kind != itemkind.Workspace {
		t.Errorf("Parent(c1) = %v", parent)
	}
	if _, found := tree.Parent("missing"); found {
		t.Error("unknown ids have no parent")
	}
	if container, found := tree.Container("f1"); !found || container.ID != "f1" {
		t.Error("folders are containers")
	}
	if _, found := tree.Container("r2"); found {
		t.Error("requests are not containers")
	}
	if _, found := tree.Container("missing"); found {
		t.Error("unknown ids are not containers")
	}
	if got := tree.SiblingEntries("r1"); !reflect.DeepEqual(got, []string{"folder.yaml"}) {
		t.Errorf("SiblingEntries(r1) = %v", got)
	}
	if got := tree.SiblingEntries("missing"); got == nil || len(got) != 0 {
		t.Errorf("SiblingEntries(missing) = %#v", got)
	}
}

func TestTreeAncestors(t *testing.T) {
	tree := sampleTree()
	ids := func(nodes []TreeNode) []string {
		result := []string{}
		for _, node := range nodes {
			result = append(result, node.ID)
		}
		return result
	}
	cases := map[string][]string{"r1": {"c1", "f1"}, "r2": {"c1"}, "c1": {}, "missing": {}, "error:x": {}}
	for id, want := range cases {
		if got := ids(tree.Ancestors(id)); !reflect.DeepEqual(got, want) {
			t.Errorf("Ancestors(%s) = %v, want %v", id, got, want)
		}
	}
}

func TestTreeSearchAddsAncestorsOnce(t *testing.T) {
	search := NewTreeSearch()
	collection := TreeNode{ID: "c1"}
	search.AddMatch("r1", []TreeNode{collection, {ID: "f1"}})
	search.AddMatch("r2", []TreeNode{collection})
	if !reflect.DeepEqual(search.MatchIDs, []string{"r1", "r2"}) || !reflect.DeepEqual(search.AncestorIDs, []string{"c1", "f1"}) {
		t.Errorf("search = %+v", search)
	}
}

func TestNewRequestNode(t *testing.T) {
	request := NewRequest("r1", "Ping")
	request.Method = "post"
	request.URL = "https://x"
	node := NewRequestNode(request, "ping.yaml")
	want := TreeNode{ID: "r1", Kind: itemkind.Request, Name: "Ping", Method: "POST", URL: "https://x", FileName: "ping.yaml"}
	if !reflect.DeepEqual(node, want) {
		t.Errorf("NewRequestNode() = %+v", node)
	}
}

func TestRequestCopyAsAndRelativeFileRefs(t *testing.T) {
	root := filepath.Join(t.TempDir(), "acme")
	request := NewRequest("r1", "Ping")
	request.Revision = "abc"
	request.Body.Binary = FileRef{Src: filepath.Join(root, "fixtures", "a.bin")}
	request.Body.FormData = []FormDataRow{
		{Src: FileRef{Src: filepath.Join(root, "b.txt")}},
		{Src: FileRef{Src: "relative/c.txt"}},
		{Src: FileRef{Src: filepath.Join(filepath.Dir(root), "outside.txt")}},
	}
	copied := request.CopyAs("r2", "  Pong ").WithFileRefsRelativeTo(root)
	if copied.ID != "r2" || copied.Name != "Pong" || copied.Revision != "" {
		t.Fatalf("CopyAs() = %+v", copied)
	}
	got := []string{copied.Body.Binary.Src, copied.Body.FormData[0].Src.Src, copied.Body.FormData[1].Src.Src, copied.Body.FormData[2].Src.Src}
	want := []string{"fixtures/a.bin", "b.txt", "relative/c.txt", filepath.Join(filepath.Dir(root), "outside.txt")}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("relative refs = %v, want %v", got, want)
	}
	if request.Body.Binary.Src == "fixtures/a.bin" || request.ID != "r1" {
		t.Error("the original request must not change")
	}
}

func TestCollectionWithSettings(t *testing.T) {
	collection := NewCollection("c1", "Billing")
	collection.Order = []string{"a.yaml"}
	settings := Collection{ID: "other", Name: "Other", Description: "docs", Auth: Auth{Type: authtype.None}, Headers: []KeyValue{NewKeyValue("X", "1")}}
	updated := collection.WithSettings(settings)
	if updated.ID != "c1" || updated.Name != "Billing" || updated.Description != "docs" || updated.Auth.Type != authtype.None {
		t.Fatalf("WithSettings() = %+v", updated)
	}
	if !reflect.DeepEqual(updated.Headers, settings.Headers) || !reflect.DeepEqual(updated.Order, []string{"a.yaml"}) {
		t.Errorf("WithSettings() = %+v", updated)
	}
	settings.Headers[0].Key = "changed"
	if updated.Headers[0].Key != "X" {
		t.Error("headers must be copied")
	}
}
