package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

func TestFolderInfoRules(t *testing.T) {
	cases := []struct {
		folder    FolderInfo
		available bool
		workspace bool
	}{
		{folder: FolderInfo{}, available: true, workspace: false},
		{folder: FolderInfo{Exists: true, IsDirectory: true}, available: true, workspace: false},
		{folder: FolderInfo{Exists: true, IsDirectory: true, EntryCount: 1, HasWorkspaceFile: true}, available: false, workspace: true},
		{folder: FolderInfo{Exists: true}, available: false, workspace: false},
	}
	for _, tc := range cases {
		if got := tc.folder.IsEmptyOrAbsent(); got != tc.available {
			t.Errorf("%+v IsEmptyOrAbsent() = %v", tc.folder, got)
		}
		if got := tc.folder.IsWorkspace(); got != tc.workspace {
			t.Errorf("%+v IsWorkspace() = %v", tc.folder, got)
		}
	}
}

func TestCompactHistoryCompactsEachSourceOnlyPastItsThreshold(t *testing.T) {
	sourceOf := func(source enums.HistorySource) enums.HistorySource { return source }
	caps := HistoryCaps{User: 100, AI: 100}
	below := append(repeatedSource(historysource.User, 120), repeatedSource(historysource.AI, 120)...)
	if kept, compact := CompactHistory(below, sourceOf, caps); compact || len(kept) != 240 {
		t.Fatalf("below threshold: %d %v", len(kept), compact)
	}
	above := append(repeatedSource(historysource.AI, 121), repeatedSource(historysource.User, 30)...)
	kept, compact := CompactHistory(above, sourceOf, caps)
	if !compact || len(kept) != 130 || kept[0] != historysource.AI || kept[len(kept)-1] != historysource.User {
		t.Fatalf("above threshold: %d %v", len(kept), compact)
	}
}

func repeatedSource(source enums.HistorySource, count int) []enums.HistorySource {
	sources := make([]enums.HistorySource, count)
	for index := range sources {
		sources[index] = source
	}
	return sources
}

func TestRequestStorageHelpers(t *testing.T) {
	request := Request{Method: "  "}
	if request.MethodOrDefault() != "GET" || request.HasID() {
		t.Fatalf("blank request: %q %v", request.MethodOrDefault(), request.HasID())
	}
	request = Request{ID: "r1", Method: " post "}
	if request.MethodOrDefault() != "POST" || !request.HasID() {
		t.Fatalf("stored request: %q %v", request.MethodOrDefault(), request.HasID())
	}
	current := RequestChange{RequestID: "r1", Revision: "b"}
	if current.BlocksSave("b", false) || !current.BlocksSave("a", false) || current.BlocksSave("a", true) {
		t.Fatal("only an unforced save from another revision is blocked")
	}
}

func TestNewBodyRawDefaultsToJson(t *testing.T) {
	if raw := NewBodyRaw("", "{}"); raw.Language != rawlanguage.Json || raw.Content != "{}" {
		t.Fatalf("NewBodyRaw(\"\") = %+v", raw)
	}
	if raw := NewBodyRaw("XML", "<a/>"); raw.Language != rawlanguage.Xml {
		t.Fatalf("NewBodyRaw(XML) = %+v", raw)
	}
}

func TestTreeNodeConstructors(t *testing.T) {
	container := NewContainerNode(NewFolder("f1", " Invoices "), "invoices", []string{"a.yaml"}).WithWarning("careful")
	if container.ID != "f1" || container.Kind != itemkind.Folder || container.Name != "Invoices" || container.FileName != "invoices" || container.Warning != "careful" || len(container.Entries) != 1 {
		t.Fatalf("NewContainerNode = %+v", container)
	}
	broken := NewErrorNode("billing/notes", "notes", "ignored", true)
	if broken.ID != "error:billing/notes" || broken.Kind != itemkind.Error || broken.Name != "notes" || !broken.Ignored || broken.Error != "ignored" {
		t.Fatalf("NewErrorNode = %+v", broken)
	}
	if NormalizeItemName("  Ping ") != "Ping" || (TreeNode{Name: " Ping"}).CopyName() != "Ping Copy" {
		t.Fatal("item names are trimmed")
	}
}
