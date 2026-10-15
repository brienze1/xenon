package persistence

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/updatefailure"
)

type tarEntry struct {
	name     string
	content  string
	linkname string
	kind     byte
}

func writeTarGz(t *testing.T, path string, entries []tarEntry) {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = file.Close() }()
	compressed := gzip.NewWriter(file)
	archive := tar.NewWriter(compressed)
	for _, entry := range entries {
		header := &tar.Header{Name: entry.name, Typeflag: entry.kind, Mode: 0o755, Size: int64(len(entry.content)), Linkname: entry.linkname}
		if entry.kind != tar.TypeReg {
			header.Size = 0
		}
		if err := archive.WriteHeader(header); err != nil {
			t.Fatal(err)
		}
		if entry.kind == tar.TypeReg {
			if _, err := archive.Write([]byte(entry.content)); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	if err := compressed.Close(); err != nil {
		t.Fatal(err)
	}
}

func linuxInstallation(t *testing.T) entity.Installation {
	t.Helper()
	folder := t.TempDir()
	path := filepath.Join(folder, "xenon")
	if err := os.WriteFile(path, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	return entity.Installation{Path: path, Platform: "linux-amd64", Installable: true}
}

func TestInstallRollsBackWhenTheSwapFails(t *testing.T) {
	installation := linuxInstallation(t)
	packagePath := filepath.Join(t.TempDir(), "xenon-0.1.7-linux-amd64.tar.gz")
	writeTarGz(t, packagePath, []tarEntry{{name: "xenon", content: "new", kind: tar.TypeReg}})
	calls := 0
	installer := &updateInstallPersistence{rename: func(from, to string) error {
		calls++
		if calls == 2 {
			return errors.New("disk full")
		}
		return os.Rename(from, to)
	}}

	err := installer.InstallUpdate(context.Background(), installation, packagePath)

	if entity.UpdateFailureOf(err, updatefailure.None).Reason != updatefailure.InstallFailed {
		t.Fatalf("expected install_failed, got %v", err)
	}
	content, readErr := os.ReadFile(installation.Path)
	if readErr != nil || string(content) != "old" {
		t.Fatalf("the previous app must be back, got %q, %v", content, readErr)
	}
	if _, statErr := os.Stat(installation.ExtractionPath()); !os.IsNotExist(statErr) {
		t.Fatal("the extraction folder must be removed")
	}
}

func TestInstallRefusesUnsafeArchives(t *testing.T) {
	cases := map[string][]tarEntry{
		"traversal":     {{name: "../escape", content: "x", kind: tar.TypeReg}},
		"absolute":      {{name: "/escape", content: "x", kind: tar.TypeReg}},
		"outside link":  {{name: "xenon", kind: tar.TypeSymlink, linkname: "../../etc/passwd"}},
		"absolute link": {{name: "xenon", kind: tar.TypeSymlink, linkname: "/etc/passwd"}},
		"two entries":   {{name: "xenon", content: "x", kind: tar.TypeReg}, {name: "extra", content: "y", kind: tar.TypeReg}},
		"wrong name":    {{name: "other", content: "x", kind: tar.TypeReg}},
	}
	for name, entries := range cases {
		t.Run(name, func(t *testing.T) {
			installation := linuxInstallation(t)
			packagePath := filepath.Join(t.TempDir(), "xenon-0.1.7-linux-amd64.tar.gz")
			writeTarGz(t, packagePath, entries)

			err := NewUpdateInstallPersistence().InstallUpdate(context.Background(), installation, packagePath)

			if entity.UpdateFailureOf(err, updatefailure.None).Reason != updatefailure.InstallFailed {
				t.Fatalf("expected install_failed, got %v", err)
			}
			if content, _ := os.ReadFile(installation.Path); string(content) != "old" {
				t.Fatalf("the app must be untouched, got %q", content)
			}
			if _, statErr := os.Stat(filepath.Join(installation.Folder(), "..", "escape")); !os.IsNotExist(statErr) {
				t.Fatal("nothing may be written outside the extraction folder")
			}
		})
	}
}

func TestInstallKeepsLinksInsideABundle(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("macOS bundles are never installed on Windows, and symbolic links need extra rights there")
	}
	folder := t.TempDir()
	installation := entity.Installation{Path: filepath.Join(folder, "Xenon.app"), Platform: "darwin-universal", Installable: true}
	packagePath := filepath.Join(t.TempDir(), "xenon-0.1.7-darwin-universal.tar.gz")
	writeTarGz(t, packagePath, []tarEntry{
		{name: "Xenon.app/Contents/Frameworks/Lib.framework/Versions/A/Lib", content: "lib", kind: tar.TypeReg},
		{name: "Xenon.app/Contents/Frameworks/Lib.framework/Lib", kind: tar.TypeSymlink, linkname: "Versions/A/Lib"},
	})

	if err := NewUpdateInstallPersistence().InstallUpdate(context.Background(), installation, packagePath); err != nil {
		t.Fatalf("a link inside the bundle is allowed: %v", err)
	}
	content, err := os.ReadFile(filepath.Join(installation.Path, "Contents", "Frameworks", "Lib.framework", "Lib"))
	if err != nil || string(content) != "lib" {
		t.Fatalf("the link must resolve inside the bundle, got %q, %v", content, err)
	}
}
