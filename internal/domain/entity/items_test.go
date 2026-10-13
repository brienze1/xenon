package entity

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
)

func TestNewRequestDefaults(t *testing.T) {
	request := NewRequest("01ID", "  Create invoice ")
	if request.ID != "01ID" || request.Name != "Create invoice" || request.Method != "GET" {
		t.Fatalf("NewRequest = %+v", request)
	}
	if request.Params == nil || request.PathVariables == nil || request.Headers == nil || request.DisabledAutoHeaders == nil {
		t.Fatal("collections must be empty, not nil")
	}
	if !request.Auth.IsInherited() || request.Body.HasContent() || !request.Settings.VerifySsl {
		t.Fatal("defaults missing")
	}
	if !request.HasValidName() || (Request{Name: " "}).HasValidName() {
		t.Error("HasValidName misbehaves")
	}
}

func TestRequestMethod(t *testing.T) {
	cases := map[string]bool{
		"get":                   true,
		" purge ":               true,
		"M-SEARCH":              true,
		"X_1":                   true,
		"1GET":                  false,
		"":                      false,
		"GE T":                  false,
		strings.Repeat("A", 32): true,
		strings.Repeat("A", 33): false,
	}
	for method, want := range cases {
		request := Request{Method: method}
		if got := request.HasValidMethod(); got != want {
			t.Errorf("HasValidMethod(%q) = %v, want %v", method, got, want)
		}
	}
	if got := (Request{Method: " purge "}).NormalizedMethod(); got != "PURGE" {
		t.Errorf("NormalizedMethod = %q", got)
	}
}

func TestRequestDisabledAutoHeaders(t *testing.T) {
	request := Request{DisabledAutoHeaders: []string{"Accept", "User-Agent"}}
	if !request.IsAutoHeaderDisabled("user-agent") || request.IsAutoHeaderDisabled("Host") {
		t.Error("IsAutoHeaderDisabled must match names case-insensitively")
	}
}

func TestCollectionAndFolder(t *testing.T) {
	collection := NewCollection("01C", " Billing API ")
	if collection.Kind != itemkind.Collection || collection.IsFolder() || collection.Name != "Billing API" {
		t.Fatalf("NewCollection = %+v", collection)
	}
	if collection.Order == nil || collection.Headers == nil || !collection.Auth.IsInherited() || collection.ID != "01C" {
		t.Fatal("collection defaults missing")
	}
	folder := NewFolder("01F", "Invoices")
	if folder.Kind != itemkind.Folder || !folder.IsFolder() {
		t.Fatalf("NewFolder = %+v", folder)
	}
	if !folder.HasValidName() || (Collection{}).HasValidName() {
		t.Error("HasValidName misbehaves")
	}
}

func TestNewWorkspace(t *testing.T) {
	workspace := NewWorkspace("01W", "  Acme ")
	if workspace.ID != "01W" || workspace.Name != "Acme" || workspace.Order == nil || workspace.Description != "" {
		t.Fatalf("NewWorkspace = %+v", workspace)
	}
	if !workspace.HasValidName() || (Workspace{Name: strings.Repeat("x", 201)}).HasValidName() {
		t.Error("HasValidName misbehaves")
	}
}

func TestRegisteredWorkspaceOverlaps(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "work", "acme")
	addedAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	registered := NewRegisteredWorkspace("01L", root+string(filepath.Separator), "Acme", addedAt)
	if registered.Path != root || registered.LocalID != "01L" || registered.DisplayNameCache != "Acme" || !registered.AddedAt.Equal(addedAt) || registered.Missing {
		t.Fatalf("NewRegisteredWorkspace = %+v", registered)
	}
	cases := map[string]bool{
		root:                                   true,
		filepath.Join(root, "nested"):          true,
		filepath.Join(root, "..", "acme"):      true,
		filepath.Dir(root):                     true,
		filepath.Join(root + "-other"):         false,
		filepath.Join(filepath.Dir(root), "x"): false,
		"relative":                             false,
	}
	for path, want := range cases {
		if got := registered.Overlaps(path); got != want {
			t.Errorf("Overlaps(%q) = %v, want %v", path, got, want)
		}
	}
}

func TestIsSameOrInside(t *testing.T) {
	parent := filepath.Join(string(filepath.Separator), "a", "b")
	if !isSameOrInside(parent, parent) || !isSameOrInside(filepath.Join(parent, "c"), parent) {
		t.Error("same and child paths are inside")
	}
	if isSameOrInside(filepath.Dir(parent), parent) || isSameOrInside(filepath.Join(parent+"c"), parent) {
		t.Error("parent and sibling paths are not inside")
	}
}

func TestAppRegistry(t *testing.T) {
	registry := NewAppRegistry()
	if registry.Version != 1 || registry.Workspaces == nil || registry.ActiveLocalID != "" {
		t.Fatalf("NewAppRegistry = %+v", registry)
	}
	root := filepath.Join(string(filepath.Separator), "work")
	if registry.OverlapsOther(filepath.Join(root, "acme"), "") {
		t.Error("empty registry overlaps nothing")
	}
	registry.Register(NewRegisteredWorkspace("01A", filepath.Join(root, "acme"), "Acme", time.Time{}))
	registry.Register(NewRegisteredWorkspace("01G", filepath.Join(root, "globex"), "Globex", time.Time{}))
	if registry.ActiveLocalID != "01G" || len(registry.Workspaces) != 2 {
		t.Fatalf("Register must append and activate, got %+v", registry)
	}
	if !registry.OverlapsOther(filepath.Join(root, "acme", "inner"), "") || registry.OverlapsOther(filepath.Join(root, "initech"), "") {
		t.Error("OverlapsOther must check every registered workspace")
	}
	found, ok := registry.Find("01A")
	if !ok || found.DisplayNameCache != "Acme" {
		t.Fatalf("Find(01A) = %+v, %v", found, ok)
	}
	if _, ok := registry.Find("missing"); ok {
		t.Error("Find(missing) must fail")
	}
	registry.Workspaces[0].Missing = true
	activated, ok := registry.Activate("01A", "Acme Corp")
	if !ok || activated != &registry.Workspaces[0] {
		t.Fatal("Activate(01A) must succeed and return the registered workspace")
	}
	if registry.ActiveLocalID != "01A" || registry.Workspaces[0].DisplayNameCache != "Acme Corp" || registry.Workspaces[0].Missing {
		t.Fatalf("Activate must refresh the cache, got %+v", registry)
	}
	if missing, ok := registry.Activate("missing", "x"); ok || missing != nil || registry.ActiveLocalID != "01A" {
		t.Error("Activate(missing) must fail and keep the active workspace")
	}
}
