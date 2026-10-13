package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureGitignoreAppendsTheXenonBlockOnce(t *testing.T) {
	cases := map[string]string{
		"":                       gitignoreBlock,
		"node_modules/":          "node_modules/\n" + gitignoreBlock,
		"node_modules/\n":        "node_modules/\n" + gitignoreBlock,
		"a\n" + gitignoreBlock:   "a\n" + gitignoreBlock,
		gitignoreHeader + "\n.x": gitignoreHeader + "\n.x",
	}
	for existing, want := range cases {
		root := t.TempDir()
		path := filepath.Join(root, gitignoreFile)
		if existing != "" {
			if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		if err := ensureGitignore(root); err != nil {
			t.Fatalf("ensureGitignore() error = %v", err)
		}
		got, _ := os.ReadFile(path)
		if string(got) != want {
			t.Errorf("existing %q: got %q, want %q", existing, got, want)
		}
	}
}

func TestEnsureGitignoreReportsUnreadableFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, gitignoreFile), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := ensureGitignore(root); err == nil {
		t.Fatal("a .gitignore folder cannot be read")
	}
}
