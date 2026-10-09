package usecase

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func codeOf(err error) string {
	if custom, ok := err.(errs.BaseError); ok {
		return custom.Code()
	}
	return ""
}

func TestValidateWorkspaceFolders(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "work")
	registry := entity.NewAppRegistry()
	registry.Register(entity.NewRegisteredWorkspace("01A", filepath.Join(root, "acme"), "Acme", time.Time{}))
	acme := entity.FolderInfo{Path: filepath.Join(root, "acme"), Exists: true, IsDirectory: true, EntryCount: 2, HasWorkspaceFile: true}
	fresh := entity.FolderInfo{Path: filepath.Join(root, "fresh")}
	full := entity.FolderInfo{Path: filepath.Join(root, "full"), Exists: true, IsDirectory: true, EntryCount: 1}
	cases := []struct {
		err  error
		name string
		want string
	}{
		{name: "new free", err: ValidateNewWorkspaceFolder(fresh, registry), want: ""},
		{name: "new overlapping", err: ValidateNewWorkspaceFolder(acme, registry), want: "WKS-02409"},
		{name: "new not empty", err: ValidateNewWorkspaceFolder(full, registry), want: "WKS-01409"},
		{name: "existing own", err: ValidateExistingWorkspaceFolder(acme, registry, "01A"), want: ""},
		{name: "existing other", err: ValidateExistingWorkspaceFolder(acme, registry, ""), want: "WKS-02409"},
		{name: "existing not a workspace", err: ValidateExistingWorkspaceFolder(full, registry, ""), want: "WKS-03404"},
	}
	for _, tc := range cases {
		if got := codeOf(tc.err); got != tc.want {
			t.Errorf("%s: code %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestResolveTreeNodes(t *testing.T) {
	tree := entity.NewTree([]entity.TreeNode{entity.NewContainerNode(entity.NewCollection("c1", "Billing"), "billing", nil)}, nil)
	if parent, err := ResolveParent(tree, "c1"); err != nil || parent.ID != "c1" {
		t.Fatalf("ResolveParent(c1) = %+v, %v", parent, err)
	}
	if _, err := ResolveParent(tree, "missing"); codeOf(err) != "TRE-01404" {
		t.Fatalf("ResolveParent(missing) = %v", err)
	}
	if item, err := ResolveItem(tree, "c1"); err != nil || item.ID != "c1" {
		t.Fatalf("ResolveItem(c1) = %+v, %v", item, err)
	}
	if _, err := ResolveItem(tree, "missing"); codeOf(err) != "TRE-02404" {
		t.Fatalf("ResolveItem(missing) = %v", err)
	}
}
