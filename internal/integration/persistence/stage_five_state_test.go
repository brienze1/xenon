package persistence

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/digestalgorithm"
	"github.com/brienze1/xenon/internal/domain/enums/digestqop"
	"github.com/brienze1/xenon/internal/domain/enums/snippetlanguage"
)

func TestDigestStateLivesBesideTheUIState(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XENON_DATA_DIR", dataDir)
	persistence := NewUIStatePersistence()
	ctx := context.Background()
	if state, err := persistence.FindDigestState(ctx, "ws-1", "owner"); err != nil || state != nil {
		t.Fatalf("FindDigestState() before any save = %+v, %v", state, err)
	}
	saved := entity.DigestState{Realm: "r", Nonce: "n", Opaque: "o", Algorithm: digestalgorithm.SHA256, Qop: digestqop.Auth, NonceCount: 3}
	if err := persistence.SaveDigestState(ctx, "ws-1", "owner", saved); err != nil {
		t.Fatal(err)
	}
	ui := entity.NewUIState()
	ui.Tabs = []entity.Tab{{TabID: "t1", RequestID: "r1"}}
	ui.ClosedTabs = []entity.ClosedTab{{ClosedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Tab: entity.Tab{TabID: "t0", RequestID: "r0"}}}
	ui.Snippets = []entity.SnippetOptions{{Language: snippetlanguage.Curl, Options: map[string]string{"multiLine": "false"}}}
	if err := persistence.Save(ctx, "ws-1", ui); err != nil {
		t.Fatal(err)
	}
	found, err := persistence.FindDigestState(ctx, "ws-1", "owner")
	if err != nil || found == nil || *found != saved {
		t.Fatalf("a UI save keeps the digest state: %+v, %v", found, err)
	}
	if other, _ := persistence.FindDigestState(ctx, "ws-1", "other"); other != nil {
		t.Fatal("another owner has no state")
	}
	state, err := persistence.Find(ctx, "ws-1")
	if err != nil || len(state.Tabs) != 1 || len(state.ClosedTabs) != 1 || state.ClosedTabs[0].Tab.TabID != "t0" || state.Snippets[0].Options["multiLine"] != "false" || state.Snippets[0].Language != snippetlanguage.Curl {
		t.Fatalf("Find() = %+v, %v", state, err)
	}
	path := filepath.Join(dataDir, workspacesDataDir, "ws-1", uiStateFile)
	if err := os.WriteFile(path, []byte("{ broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	if state, err := persistence.FindDigestState(ctx, "ws-1", "owner"); err != nil || state != nil {
		t.Fatalf("a broken file has no digest state: %+v, %v", state, err)
	}
	if err := persistence.SaveDigestState(ctx, "ws-1", "owner", saved); err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	data, _ := os.ReadFile(path)
	if json.Unmarshal(data, &fields) != nil || fields[digestStateKey] == nil {
		t.Fatalf("the broken file is replaced: %s", data)
	}
	if os.PathSeparator != '\\' && os.Getuid() != 0 && os.Chmod(path, 0o000) == nil {
		if _, err := persistence.FindDigestState(ctx, "ws-1", "owner"); errorCode(err) != "UIS-09500" {
			t.Fatalf("an unreadable file = %v", err)
		}
		if err := persistence.SaveDigestState(ctx, "ws-1", "owner", saved); errorCode(err) != "UIS-09500" {
			t.Fatalf("an unreadable file = %v", err)
		}
		_ = os.Chmod(path, 0o600)
	}
}

func TestDigestStateWriteFailures(t *testing.T) {
	dataDir := t.TempDir()
	t.Setenv("XENON_DATA_DIR", dataDir)
	if err := os.WriteFile(filepath.Join(dataDir, workspacesDataDir), []byte("file"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := NewUIStatePersistence().SaveDigestState(context.Background(), "ws-1", "o", entity.DigestState{})
	if code := errorCode(err); code != "UIS-09500" && code != "UIS-07500" {
		t.Fatalf("SaveDigestState() = %v (%s)", err, code)
	}
}

func TestTokenLocksSerialiseOneOwner(t *testing.T) {
	persistence := NewVaultPersistence(nil)
	ctx := context.Background()
	unlock := persistence.LockToken(ctx, "ws", "owner")
	other := persistence.LockToken(ctx, "ws", "other")
	other()
	acquired := make(chan struct{})
	var once sync.Once
	go func() {
		release := persistence.LockToken(ctx, "ws", "owner")
		once.Do(func() { close(acquired) })
		release()
	}()
	select {
	case <-acquired:
		t.Fatal("the owner lock is held")
	case <-time.After(30 * time.Millisecond):
	}
	unlock()
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("the owner lock is released")
	}
}
