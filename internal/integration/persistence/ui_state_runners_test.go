package persistence

import (
	"context"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func TestRunnerPreferencesSurviveAUIStateSaveAndCanBeDeleted(t *testing.T) {
	t.Setenv("XENON_DATA_DIR", t.TempDir())
	persistence := NewUIStatePersistence()
	ctx := context.Background()
	config := entity.NewRunConfig()
	config.Rows = []entity.RunRow{{RequestID: "logout", Selected: true}}
	if err := persistence.SaveRunnerPreferences(ctx, "ws-1", "shop", entity.NewRunnerPreferences(config)); err != nil {
		t.Fatalf("SaveRunnerPreferences() error = %v", err)
	}
	if err := persistence.SaveRunnerPreferences(ctx, "ws-1", "orders", entity.NewRunnerPreferences(entity.NewRunConfig())); err != nil {
		t.Fatalf("SaveRunnerPreferences() error = %v", err)
	}
	if err := persistence.Save(ctx, "ws-1", entity.NewUIState()); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	preferences, err := persistence.FindRunnerPreferences(ctx, "ws-1")
	if err != nil || len(preferences) != 2 || preferences["shop"].Config.Rows[0].RequestID != "logout" {
		t.Fatalf("FindRunnerPreferences() = %+v, %v", preferences, err)
	}
	state, err := persistence.Find(ctx, "ws-1")
	if err != nil || len(state.Runners) != 2 {
		t.Fatalf("Find() runners = %+v, %v", state, err)
	}
	if err := persistence.DeleteRunnerPreferences(ctx, "ws-1", []string{"shop", "orders"}); err != nil {
		t.Fatalf("DeleteRunnerPreferences() error = %v", err)
	}
	if preferences, _ := persistence.FindRunnerPreferences(ctx, "ws-1"); len(preferences) != 0 {
		t.Errorf("preferences after delete = %+v", preferences)
	}
}
