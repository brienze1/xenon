package adapter

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

type zipEntry struct {
	name    string
	content string
	mode    os.FileMode
}

func writeZip(t *testing.T, entries ...zipEntry) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "source.zip")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := zip.NewWriter(file)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate}
		header.SetMode(entry.mode)
		part, err := writer.CreateHeader(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write([]byte(entry.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func extractCode(t *testing.T, archive string, limit int64) (string, string, string) {
	t.Helper()
	target := t.TempDir()
	err := NewArchive().Extract(archive, target, limit)
	if err == nil {
		return "", "", target
	}
	custom := err.(errs.BaseError)
	return custom.Code(), custom.Description(), target
}

func TestExtractUnpacksFoldersAndFiles(t *testing.T) {
	archive := writeZip(t, zipEntry{"export/", "", os.ModeDir | 0o755}, zipEntry{"export/archive.json", "{}", 0o644}, zipEntry{`export\collection\a.json`, "[1]", 0o644})
	code, _, target := extractCode(t, archive, 5)
	if code != "" {
		t.Fatalf("Extract() code = %s", code)
	}
	for path, want := range map[string]string{"export/archive.json": "{}", "export/collection/a.json": "[1]"} {
		got, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(path)))
		if err != nil || string(got) != want {
			t.Errorf("%s = %q, %v", path, got, err)
		}
	}
}

func TestExtractRefusesUnsafeEntriesBeforeWritingAnything(t *testing.T) {
	for _, entry := range []zipEntry{
		{"../evil.json", "{}", 0o644},
		{`a\..\..\evil.json`, "{}", 0o644},
		{"/tmp/evil.json", "{}", 0o644},
		{"link.json", "/etc/passwd", os.ModeSymlink | 0o777},
	} {
		archive := writeZip(t, zipEntry{"ok.json", "{}", 0o644}, entry)
		code, message, target := extractCode(t, archive, 1000)
		if code != "IMP-03400" || message != "source.zip contains the unsafe entry "+entry.name {
			t.Errorf("%s: %s %q", entry.name, code, message)
		}
		if listing, _ := os.ReadDir(target); len(listing) != 0 {
			t.Errorf("%s: wrote %d entries", entry.name, len(listing))
		}
	}
}

func TestExtractStopsAtTheLimit(t *testing.T) {
	archive := writeZip(t, zipEntry{"a.json", strings.Repeat("a", 6), 0o644}, zipEntry{"b.json", strings.Repeat("b", 5), 0o644})
	if code, message, _ := extractCode(t, archive, 10); code != "IMP-02413" || message != "source.zip unpacks to more than 10 bytes" {
		t.Fatalf("limit = %s %q", code, message)
	}
	if code, _, _ := extractCode(t, archive, 11); code != "" {
		t.Fatalf("exactly the limit is allowed, got %s", code)
	}
}

func TestExtractRefusesBrokenArchivesAndUnwritableTargets(t *testing.T) {
	broken := filepath.Join(t.TempDir(), "broken.zip")
	if err := os.WriteFile(broken, []byte("not a zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	if code, message, _ := extractCode(t, broken, 10); code != "IMP-18400" || message != "broken.zip is not a valid zip archive" {
		t.Fatalf("broken = %s %q", code, message)
	}
	archive := writeZip(t, zipEntry{"dir/a.json", "{}", 0o644})
	blocked := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocked, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := NewArchive().Extract(archive, blocked, 10); err == nil || err.(errs.BaseError).Code() != "IMP-19500" {
		t.Fatalf("unwritable target = %v", err)
	}
}
