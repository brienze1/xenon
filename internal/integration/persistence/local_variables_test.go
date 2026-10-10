package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestLocalVariablesRoundTripWithOwnerOnlyPermissions(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XENON_DATA_DIR", dataDir)
	persistence := NewLocalVariablesPersistence()
	ctx := context.Background()
	if _, err := persistence.Find(ctx, "ws-1"); !errs.IsNotFoundError(err) || errorCode(err) != "VAR-03500" {
		t.Fatalf("Find() before any save error = %v", err)
	}
	local := entity.NewLocalVariables().WithActiveEnvironment("01E")
	local.Current["env:01E"] = map[string]string{"base_url": "http://localhost"}
	if err := persistence.Save(ctx, "ws-1", local); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	path := filepath.Join(dataDir, workspacesDataDir, "ws-1", localVariablesFile)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	dirInfo, _ := os.Stat(filepath.Dir(path))
	if skipPermissions := os.PathSeparator == '\\'; !skipPermissions && (info.Mode().Perm() != 0o600 || dirInfo.Mode().Perm() != 0o700) {
		t.Errorf("permissions file %o dir %o", info.Mode().Perm(), dirInfo.Mode().Perm())
	}
	found, err := persistence.Find(ctx, "ws-1")
	if err != nil || found.ActiveEnvironmentID != "01E" || found.Current["env:01E"]["base_url"] != "http://localhost" {
		t.Fatalf("Find() = %+v, %v", found, err)
	}
	if err := os.WriteFile(path, []byte("{ not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := persistence.Find(ctx, "ws-1"); !errs.IsUnprocessableEntityError(err) || errorCode(err) != "VAR-04422" {
		t.Fatalf("Find() of invalid JSON error = %v", err)
	}
}

func TestLocalVariablesSaveReportsFileErrors(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XENON_DATA_DIR", dataDir)
	persistence := NewLocalVariablesPersistence()
	writeTestFile(t, filepath.Join(dataDir, workspacesDataDir, "as-file"), "x")
	if err := persistence.Save(context.Background(), "as-file", entity.NewLocalVariables()); errorCode(err) != "VAR-05500" {
		t.Fatalf("Save() under a file error = %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dataDir, workspacesDataDir, "ws", localVariablesFile), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := persistence.Save(context.Background(), "ws", entity.NewLocalVariables()); errorCode(err) != "VAR-06500" {
		t.Fatalf("Save() over a folder error = %v", err)
	}
}
