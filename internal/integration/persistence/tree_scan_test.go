package persistence

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/pkg/timeutils"
)

func TestReadDocumentReasons(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return path
	}
	cases := map[string]string{
		write("ok.yaml", "xenon: 1\nkind: request\nname: A\n"):      "",
		write("nokind.yaml", "name: A\n"):                           "",
		write("future.yaml", "xenon: 2\nkind: request\n"):           "future.yaml was written by a newer Xenon (format 2)",
		write("folder.yaml", "xenon: 1\nkind: folder\n"):            "folder.yaml is a folder file, expected a request",
		write("broken.yaml", "name: [x\n"):                          "broken.yaml cannot be read: ",
		write("typed.yaml", "xenon: one\n"):                         "typed.yaml cannot be read: ",
		filepath.Join(dir, "missing.yaml"):                          "missing.yaml cannot be read: ",
		write("current.yaml", "xenon: 1\nkind: REQUEST\nname: A\n"): "",
	}
	for path, want := range cases {
		_, reason := readRequest(path)
		if (want == "" && reason != "") || !strings.HasPrefix(reason, want) {
			t.Errorf("%s: reason %q, want prefix %q", filepath.Base(path), reason, want)
		}
	}
}

func TestIsOwnWrite(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	timeutils.SetCurrentTimeProvider(func() time.Time { return now })
	defer timeutils.SetCurrentTimeProvider(time.Now)
	ctx := context.Background()
	index := NewWorkspaceIndexCache()
	index.RecordWrite(ctx, "/a.yaml", []byte("one"))
	if !index.IsOwnWrite(ctx, "/a.yaml", []byte("one")) {
		t.Fatal("a fresh write with the same content is ours")
	}
	if index.IsOwnWrite(ctx, "/a.yaml", []byte("two")) || index.IsOwnWrite(ctx, "/b.yaml", []byte("one")) {
		t.Fatal("other content or paths are not ours")
	}
	now = now.Add(2 * time.Second)
	if !index.IsOwnWrite(ctx, "/a.yaml", []byte("one")) {
		t.Fatal("still ours at the end of the window")
	}
	now = now.Add(time.Millisecond)
	if index.IsOwnWrite(ctx, "/a.yaml", []byte("one")) {
		t.Fatal("not ours after the window")
	}
}

func TestMoveItemRenamesThroughATemporaryNameWhenOnlyCaseChanges(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "Users.yaml")
	if err := os.WriteFile(source, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := moveItem(source, source); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "users.yaml")
	if err := moveItem(source, target); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 || entries[0].Name() != "users.yaml" {
		t.Fatalf("entries = %v", entries)
	}
	other := filepath.Join(dir, "other.yaml")
	if err := os.WriteFile(other, []byte("y"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := moveItem(other, target); err == nil {
		t.Fatal("moving onto an existing file must fail")
	}
	if err := moveItem(filepath.Join(dir, "missing.yaml"), filepath.Join(dir, "new.yaml")); err == nil {
		t.Fatal("moving a missing file must fail")
	}
}
