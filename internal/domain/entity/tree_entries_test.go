package entity

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func entriesTree() Tree {
	created := TreeNode{ID: "r1", Kind: itemkind.Request, Name: "Create invoice", Method: "POST", Examples: []TreeExample{{ID: "e1", Name: "Created", Code: 201}, {ID: "e2", Name: "Rejected", Code: 400}}}
	folder := TreeNode{ID: "f1", Kind: itemkind.Folder, Name: "Invoices", Children: []TreeNode{created}}
	ping := TreeNode{ID: "r2", Kind: itemkind.Request, Name: "Ping", Method: "GET"}
	billing := TreeNode{ID: "c1", Kind: itemkind.Collection, Name: "Billing", Children: []TreeNode{folder, ping}}
	users := TreeNode{ID: "c2", Kind: itemkind.Collection, Name: "Users"}
	return NewTree([]TreeNode{billing, users}, []string{"billing", "users"})
}

func TestTreeEntriesArePreOrderWithExamplesBeforeChildren(t *testing.T) {
	type line struct {
		id, kind, path, parent, request, method string
		children                                int
	}
	var got []line
	for _, entry := range entriesTree().Entries() {
		got = append(got, line{entry.ID(), entry.Kind().String(), entry.Path, entry.ParentID, entry.RequestID, entry.Method(), entry.Children()})
	}
	want := []line{
		{"c1", "collection", "Billing", "", "", "", 2},
		{"f1", "folder", "Billing / Invoices", "c1", "", "", 1},
		{"r1", "request", "Billing / Invoices / Create invoice", "f1", "", "POST", 2},
		{"e1", "example", "Billing / Invoices / Create invoice / Created", "r1", "r1", "", 0},
		{"e2", "example", "Billing / Invoices / Create invoice / Rejected", "r1", "r1", "", 0},
		{"r2", "request", "Billing / Ping", "c1", "", "GET", 0},
		{"c2", "collection", "Users", "", "", "", 0},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Entries() =\n%v\nwant\n%v", got, want)
	}
	if len(NewTree(nil, nil).Entries()) != 0 {
		t.Error("an empty tree has no entries")
	}
}

func TestTreeEntryFindsItemsAndTheirKinds(t *testing.T) {
	tree := entriesTree()
	cases := []struct {
		id         string
		found      bool
		example    bool
		request    bool
		container  bool
		collection bool
		name       string
	}{
		{id: "c1", found: true, container: true, collection: true, name: "Billing"},
		{id: "f1", found: true, container: true, name: "Invoices"},
		{id: "r1", found: true, request: true, name: "Create invoice"},
		{id: "e2", found: true, example: true, name: "Rejected"},
		{id: "missing"},
	}
	for _, tc := range cases {
		entry, found := tree.Entry(tc.id)
		if found != tc.found || entry.IsExample() != tc.example || entry.IsRequest() != tc.request || entry.IsContainer() != tc.container || entry.IsCollection() != tc.collection || entry.Name() != tc.name {
			t.Errorf("Entry(%q) = %+v %v", tc.id, entry, found)
		}
	}
	entry, _ := tree.Entry("e1")
	if entry.Target() != (AIActivityTarget{ID: "e1", Name: "Created", Path: "Billing / Invoices / Create invoice / Created"}) {
		t.Errorf("Target() = %+v", entry.Target())
	}
}

func TestTreeSearchKeepsTreeOrderAndLimit(t *testing.T) {
	tree := entriesTree()
	cases := []struct {
		name     string
		query    string
		matchIDs []string
		limit    int
		want     []string
	}{
		{name: "matched requests", query: "zzz", matchIDs: []string{"r2", "r1"}, limit: 50, want: []string{"r1", "r2"}},
		{name: "names ignoring case", query: " INVOICE ", limit: 50, want: []string{"f1", "r1"}},
		{name: "examples and collections", query: "e", limit: 50, want: []string{"f1", "r1", "e1", "e2", "c2"}},
		{name: "limit", query: "e", limit: 2, want: []string{"f1", "r1"}},
		{name: "no match", query: "nothing", limit: 50, want: []string{}},
	}
	for _, tc := range cases {
		got := []string{}
		for _, entry := range tree.Search(tc.query, tc.matchIDs, tc.limit) {
			got = append(got, entry.ID())
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: Search() = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestTreeContainerAtFindsCollectionsAndFoldersByPath(t *testing.T) {
	tree := entriesTree()
	for path, want := range map[string]string{"Billing": "c1", "Billing / Invoices": "f1", "Billing/Invoices": "f1", " Users ": "c2"} {
		if entry, found := tree.ContainerAt(path); !found || entry.ID() != want {
			t.Errorf("ContainerAt(%q) = %v, %v; want %s", path, entry.ID(), found, want)
		}
	}
	for _, path := range []string{"Billing / Ping", "Nope", ""} {
		if _, found := tree.ContainerAt(path); found {
			t.Errorf("ContainerAt(%q) must not find a container", path)
		}
	}
}
