package persistence

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/pkg/timeutils"
)

func TestEffectiveOrder(t *testing.T) {
	cases := []struct {
		listed  []string
		present []string
		want    []string
	}{
		{listed: []string{"b", "gone", "a"}, present: []string{"a", "d", "b", "c"}, want: []string{"b", "a", "c", "d"}},
		{listed: []string{"a", "a"}, present: []string{"a"}, want: []string{"a"}},
		{listed: nil, present: []string{"z", "y"}, want: []string{"y", "z"}},
		{listed: []string{"x"}, present: []string{}, want: []string{}},
	}
	for _, tc := range cases {
		if got := effectiveOrder(tc.listed, tc.present); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("effectiveOrder(%v, %v) = %v, want %v", tc.listed, tc.present, got, tc.want)
		}
	}
}

func TestIdentifyKeepsFirstIdAndReidentifiesDuplicatesStably(t *testing.T) {
	scan := &treeScan{previous: map[string]string{"/b": "kept"}, seen: map[string]string{}, reidentified: map[string]string{}}
	if id, warning := scan.identify("same", "/a"); id != "same" || warning != "" {
		t.Fatalf("first = %q %q", id, warning)
	}
	if id, warning := scan.identify("same", "/b"); id != "kept" || warning != duplicateWarning {
		t.Fatalf("duplicate = %q %q", id, warning)
	}
	id, warning := scan.identify("same", "/c")
	if id == "same" || id == "kept" || warning != duplicateWarning || scan.reidentified["/c"] != id {
		t.Fatalf("new duplicate = %q %q", id, warning)
	}
	missing, warning := scan.identify("", "/d")
	if missing == "" || warning != missingIdWarning {
		t.Fatalf("missing id = %q %q", missing, warning)
	}
	scan.previous["/e"] = "kept"
	if id, _ := scan.identify("", "/e"); id == "kept" {
		t.Fatal("a previous id already in use must not be reused")
	}
}

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
	index := NewWorkspaceIndex()
	index.recordWrite("/a.yaml", []byte("one"))
	if !index.IsOwnWrite("/a.yaml", []byte("one")) {
		t.Fatal("a fresh write with the same content is ours")
	}
	if index.IsOwnWrite("/a.yaml", []byte("two")) || index.IsOwnWrite("/b.yaml", []byte("one")) {
		t.Fatal("other content or paths are not ours")
	}
	now = now.Add(ownWriteWindow)
	if !index.IsOwnWrite("/a.yaml", []byte("one")) {
		t.Fatal("still ours at the end of the window")
	}
	now = now.Add(time.Millisecond)
	if index.IsOwnWrite("/a.yaml", []byte("one")) {
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
