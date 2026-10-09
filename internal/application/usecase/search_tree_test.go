package usecase

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func TestSearchTreeReturnsMatchesWithTheirAncestors(t *testing.T) {
	invoice := entity.TreeNode{ID: "r1", Kind: itemkind.Request, Name: "Create invoice", URL: "https://x/invoices"}
	ping := entity.TreeNode{ID: "r2", Kind: itemkind.Request, Name: "Ping", URL: "https://x/ping"}
	draft := entity.TreeNode{ID: "r3", Kind: itemkind.Request, Name: "Draft invoice"}
	folder := entity.TreeNode{ID: "f1", Kind: itemkind.Folder, Name: "Invoices", Children: []entity.TreeNode{invoice, draft}}
	billing := entity.TreeNode{ID: "c1", Kind: itemkind.Collection, Name: "Billing", Children: []entity.TreeNode{folder, ping}}
	other := entity.TreeNode{ID: "c2", Kind: itemkind.Collection, Name: "Invoice archive"}
	tree := entity.NewTree([]entity.TreeNode{billing, other}, nil)

	cases := map[string]entity.TreeSearch{
		"INVOICE": {MatchIDs: []string{"r1", "r3"}, AncestorIDs: []string{"c1", "f1"}},
		"ping":    {MatchIDs: []string{"r2"}, AncestorIDs: []string{"c1"}},
		"zzz":     {MatchIDs: []string{}, AncestorIDs: []string{}},
	}
	for query, want := range cases {
		if got := SearchTree(tree, query); !reflect.DeepEqual(got, want) {
			t.Errorf("SearchTree(%q) = %+v, want %+v", query, got, want)
		}
	}
}
