package usecase

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func TestRankQuickSearchOrdersByRankThenTreeOrder(t *testing.T) {
	create := entity.TreeNode{ID: "r1", Kind: itemkind.Request, Name: "Create invoice", Method: "POST", URL: "https://x/v1/invoices"}
	pay := entity.TreeNode{ID: "r2", Kind: itemkind.Request, Name: "Pay", URL: "https://x/v1/pay"}
	totals := entity.TreeNode{ID: "r3", Kind: itemkind.Request, Name: "Invoice totals", URL: "https://x/totals"}
	list := entity.TreeNode{ID: "r4", Kind: itemkind.Request, Name: "List", URL: "https://x/users?with=invoice"}
	folder := entity.TreeNode{ID: "f1", Kind: itemkind.Folder, Name: "Invoices", Children: []entity.TreeNode{create, pay}}
	billing := entity.TreeNode{ID: "c1", Kind: itemkind.Collection, Name: "Billing", Children: []entity.TreeNode{folder, totals, list}}
	broken := entity.TreeNode{ID: "error:x", Kind: itemkind.Error, Name: "invoice.yaml"}
	tree := entity.NewTree([]entity.TreeNode{billing, broken}, nil)

	cases := []struct {
		query string
		limit int
		want  []string
	}{
		{query: "INVOICE", want: []string{"r3", "r1", "r4", "r2"}},
		{query: "invoice", limit: 2, want: []string{"r3", "r1"}},
		{query: "invoice", limit: 1, want: []string{"r3"}},
		{query: "billing / invoices", want: []string{"r1", "r2"}},
		{query: "  ", want: []string{}},
		{query: "zzz", want: []string{}},
	}
	for _, tc := range cases {
		ids := make([]string, 0)
		for _, hit := range RankQuickSearch(tree, tc.query, tc.limit) {
			ids = append(ids, hit.ID)
		}
		if !reflect.DeepEqual(ids, tc.want) {
			t.Errorf("RankQuickSearch(%q, %d) = %v, want %v", tc.query, tc.limit, ids, tc.want)
		}
	}
	hit := RankQuickSearch(tree, "create", 0)[0]
	if hit.Path != "Billing / Invoices" || hit.Method != "POST" || hit.URL != create.URL || hit.Name != create.Name || hit.Rank != 0 {
		t.Errorf("unexpected hit %+v", hit)
	}
}

func TestQuickSearchLimitDefaultsWhenNotPositive(t *testing.T) {
	for limit, want := range map[int]int{-3: entity.DefaultQuickSearchLimit, 0: entity.DefaultQuickSearchLimit, 1: 1, 7: 7} {
		if got := entity.QuickSearchLimit(limit); got != want {
			t.Errorf("QuickSearchLimit(%d) = %d, want %d", limit, got, want)
		}
	}
}
