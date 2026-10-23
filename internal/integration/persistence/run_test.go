package persistence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/runstatus"
	"github.com/brienze1/xenon/pkg/errs"
)

func storedRunFor(target, id string, keys ...string) entity.Run {
	rows := make([]entity.RunRowResult, 0, len(keys))
	for index, key := range keys {
		rows = append(rows, entity.RunRowResult{Index: index, ResponseKey: key})
	}
	return entity.Run{ID: id, Status: runstatus.Completed, Target: entity.RunTarget{LocalID: "ws-1", ItemID: target}, Config: entity.NewRunConfig(), Rows: rows}
}

func TestRunsAreCappedPerTargetAndPrunedRunsAreReturned(t *testing.T) {
	t.Setenv("XENON_DATA_DIR", t.TempDir())
	persistence := NewRunPersistence()
	ctx := context.Background()
	for index := range storedRunsLimit {
		if _, err := persistence.SaveRun(ctx, storedRunFor("shop", fmt.Sprintf("run-%02d", index), "key")); err != nil {
			t.Fatalf("SaveRun() error = %v", err)
		}
	}
	if _, err := persistence.SaveRun(ctx, storedRunFor("orders", "run-00")); err != nil {
		t.Fatal(err)
	}
	pruned, err := persistence.SaveRun(ctx, storedRunFor("shop", "run-99"))
	if err != nil || len(pruned) != 1 || pruned[0].ID != "run-00" || len(pruned[0].ResponseKeys()) != 1 {
		t.Fatalf("pruned = %+v, %v", pruned, err)
	}
	runs, err := persistence.FindRuns(ctx, "ws-1", "shop")
	if err != nil || len(runs) != storedRunsLimit || runs[0].ID != "run-99" {
		t.Fatalf("FindRuns() = %d runs, %v", len(runs), err)
	}
	if folder, _ := persistence.FindRuns(ctx, "ws-1", "orders"); len(folder) != 1 {
		t.Errorf("a folder keeps its own runs: %d", len(folder))
	}
}

func TestStoredRunsCanBeReadDeletedClearedAndOrphaned(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XENON_DATA_DIR", dataDir)
	persistence := NewRunPersistence()
	ctx := context.Background()
	for _, run := range []entity.Run{storedRunFor("shop", "a", "run:a:0"), storedRunFor("shop", "b"), storedRunFor("gone", "c", "run:c:0"), storedRunFor("orders", "d")} {
		if _, err := persistence.SaveRun(ctx, run); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dataDir, workspacesDataDir, "ws-1", runsDataDir, "shop", "0000.json"), []byte("not a run"), 0o600); err != nil {
		t.Fatal(err)
	}
	if runs, _ := persistence.FindRuns(ctx, "ws-1", "shop"); len(runs) != 2 {
		t.Errorf("an unreadable file is left out: %d", len(runs))
	}
	if data, err := persistence.FindRunFile(ctx, "ws-1", "a"); err != nil || len(data) == 0 {
		t.Errorf("FindRunFile() = %d bytes, %v", len(data), err)
	}
	deleted, err := persistence.DeleteRun(ctx, "ws-1", "a")
	if err != nil || deleted.ID != "a" || len(deleted.ResponseKeys()) != 1 {
		t.Fatalf("DeleteRun() = %+v, %v", deleted, err)
	}
	if _, err := persistence.FindRun(ctx, "ws-1", "a"); !errs.IsNotFoundError(err) {
		t.Errorf("a deleted run is unknown: %v", err)
	}
	if _, err := persistence.DeleteRun(ctx, "ws-1", "a"); !errs.IsNotFoundError(err) {
		t.Errorf("deleting twice is unknown: %v", err)
	}
	orphans, err := persistence.DeleteOrphanRuns(ctx, "ws-1", []string{"shop", "orders"})
	if err != nil || len(orphans) != 1 || orphans[0].ID != "c" {
		t.Fatalf("DeleteOrphanRuns() = %+v, %v", orphans, err)
	}
	cleared, err := persistence.ClearRuns(ctx, "ws-1", "shop")
	if err != nil || len(cleared) != 2 {
		t.Fatalf("ClearRuns() = %+v, %v", cleared, err)
	}
	if err := persistence.DeleteWorkspaceRuns(ctx, "ws-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dataDir, workspacesDataDir, "ws-1", runsDataDir)); !os.IsNotExist(err) {
		t.Errorf("the runs folder is gone: %v", err)
	}
	if orphans, err := persistence.DeleteOrphanRuns(ctx, "ws-1", nil); err != nil || len(orphans) != 0 {
		t.Errorf("no runs folder has no orphans: %v, %v", orphans, err)
	}
}
