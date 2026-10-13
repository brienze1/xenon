package usecase

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func TestDeriveFileName(t *testing.T) {
	cases := []struct {
		name     string
		kind     enums.ItemKind
		want     string
		siblings []string
	}{
		{name: "Create invoice", kind: itemkind.Request, want: "create-invoice.yaml"},
		{name: "Billing API", kind: itemkind.Collection, want: "billing-api"},
		{name: "Invoices", kind: itemkind.Folder, want: "invoices"},
		{name: "GET /users/:id?", kind: itemkind.Request, want: "get-users-id.yaml"},
		{name: "Create invoice", kind: itemkind.Request, siblings: []string{"create-invoice.yaml"}, want: "create-invoice-2.yaml"},
		{name: "Create invoice", kind: itemkind.Request, siblings: []string{"Create-Invoice.YAML", "create-invoice-2.yaml"}, want: "create-invoice-3.yaml"},
		{name: "Invoices", kind: itemkind.Folder, siblings: []string{"INVOICES"}, want: "invoices-2"},
		{name: "collection", kind: itemkind.Request, want: "collection-2.yaml"},
		{name: "Folder", kind: itemkind.Request, want: "folder-2.yaml"},
		{name: "collection", kind: itemkind.Collection, want: "collection"},
		{name: "", kind: itemkind.Request, want: "untitled.yaml"},
		{name: "con", kind: itemkind.Folder, want: "con_"},
	}
	for _, tc := range cases {
		if got := DeriveFileName(tc.name, tc.kind, tc.siblings); got != tc.want {
			t.Errorf("DeriveFileName(%q, %s, %v) = %q, want %q", tc.name, tc.kind, tc.siblings, got, tc.want)
		}
	}
}

func TestDeriveFileNameDoesNotMutateSiblings(t *testing.T) {
	siblings := make([]string, 1, 4)
	siblings[0] = "a.yaml"
	DeriveFileName("a", itemkind.Request, siblings)
	if extended := siblings[:2]; extended[1] != "" {
		t.Fatalf("siblings backing array was modified: %v", extended)
	}
}
