package persistence

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestSaveFileWritesAtomically(t *testing.T) {
	target := filepath.Join(t.TempDir(), "out.json")
	if err := NewExportFilePersistence().SaveFile(context.Background(), target, []byte("{}")); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(target)
	if err != nil || string(content) != "{}" {
		t.Fatalf("content = %q, %v", content, err)
	}
	expectOnly(t, filepath.Dir(target), "out.json")
}

func TestSaveArchiveStreamsPortableEntries(t *testing.T) {
	target := filepath.Join(t.TempDir(), "out.zip")
	entries := []entity.ExportEntry{
		entity.NewExportEntry(`collections\café.json`, []byte("a")),
		entity.NewExportEntry("globals.json", []byte("b")),
	}
	written := make([]int, 0)
	if err := NewExportFilePersistence().SaveArchive(context.Background(), target, entries, func(done int) { written = append(written, done) }); err != nil {
		t.Fatal(err)
	}
	if len(written) != 2 || written[0] != 1 || written[1] != 2 {
		t.Errorf("written = %v", written)
	}
	reader, err := zip.OpenReader(target)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = reader.Close() }()
	names := make([]string, 0)
	for _, file := range reader.File {
		names = append(names, file.Name)
		if file.Method != zip.Deflate || file.Flags&zipUtf8NamesFlag == 0 {
			t.Errorf("%s: method %d flags %x", file.Name, file.Method, file.Flags)
		}
		opened, _ := file.Open()
		content, _ := io.ReadAll(opened)
		_ = opened.Close()
		if len(content) != 1 {
			t.Errorf("%s: content %q", file.Name, content)
		}
	}
	if strings.Join(names, ",") != "collections/café.json,globals.json" {
		t.Errorf("names = %q", names)
	}
	expectOnly(t, filepath.Dir(target), "out.zip")
}

func TestExportWritesLeaveNoFileWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	folder := t.TempDir()
	persistence := NewExportFilePersistence()
	expectCode(t, persistence.SaveFile(ctx, filepath.Join(folder, "out.json"), []byte("{}")), "EXP-05499")
	expectCode(t, persistence.SaveArchive(ctx, filepath.Join(folder, "out.zip"), []entity.ExportEntry{entity.NewExportEntry("a.json", nil)}, func(int) {}), "EXP-05499")
	expectOnly(t, folder)
}

func TestSaveArchiveStopsWhenCancelledPartWay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	folder := t.TempDir()
	entries := []entity.ExportEntry{entity.NewExportEntry("a.json", nil), entity.NewExportEntry("b.json", nil), entity.NewExportEntry("c.json", nil)}
	written := make([]int, 0)
	err := NewExportFilePersistence().SaveArchive(ctx, filepath.Join(folder, "out.zip"), entries, func(done int) {
		written = append(written, done)
		if done == 1 {
			cancel()
		}
	})
	expectCode(t, err, "EXP-05499")
	if len(written) != 1 {
		t.Errorf("written = %v", written)
	}
	expectOnly(t, folder)
}

func TestExportWritesReportFailures(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing")
	persistence := NewExportFilePersistence()
	expectCode(t, persistence.SaveFile(context.Background(), filepath.Join(missing, "out.json"), []byte("{}")), "EXP-02500")
	expectCode(t, persistence.SaveArchive(context.Background(), filepath.Join(missing, "out.zip"), nil, func(int) {}), "EXP-02500")
}

func expectOnly(t *testing.T, folder string, names ...string) {
	t.Helper()
	entries, err := os.ReadDir(folder)
	if err != nil {
		t.Fatal(err)
	}
	found := make([]string, 0, len(entries))
	for _, entry := range entries {
		found = append(found, entry.Name())
	}
	sort.Strings(found)
	if strings.Join(found, ",") != strings.Join(names, ",") {
		t.Errorf("%s holds %v, want %v", folder, found, names)
	}
}
