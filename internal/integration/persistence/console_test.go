package persistence

import (
	"context"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/consoleentrykind"
	"github.com/brienze1/xenon/internal/domain/enums/consolefilter"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestConsolePersistence(t *testing.T) {
	ctx := context.Background()
	console := NewConsolePersistence()
	dropped := console.Append(ctx, []entity.ConsoleEntry{{ID: "a", Kind: consoleentrykind.Log}, {ID: "b", Kind: consoleentrykind.Request}})
	if len(dropped) != 0 {
		t.Fatalf("dropped = %+v", dropped)
	}
	if found := console.FindEntries(ctx, entity.ConsoleQuery{Limit: 10, Filter: consolefilter.Requests}); len(found) != 1 || found[0].ID != "b" {
		t.Errorf("FindEntries = %+v", found)
	}
	if entry, err := console.FindEntry(ctx, "a"); err != nil || entry.ID != "a" {
		t.Errorf("FindEntry(a) = %+v, %v", entry, err)
	}
	if _, err := console.FindEntry(ctx, "gone"); !errs.IsNotFoundError(err) || errorCode(err) != "CON-01404" {
		t.Errorf("FindEntry(gone) error = %v", err)
	}
	full := make([]entity.ConsoleEntry, entity.ConsoleCapacity)
	if dropped := console.Append(ctx, full); len(dropped) != 2 || dropped[0].ID != "a" || dropped[1].ID != "b" {
		t.Errorf("dropped = %d", len(dropped))
	}
	if removed := console.Clear(ctx); len(removed) != entity.ConsoleCapacity {
		t.Errorf("removed = %d", len(removed))
	}
	if found := console.FindEntries(ctx, entity.ConsoleQuery{Limit: 10}); len(found) != 0 {
		t.Errorf("after Clear = %+v", found)
	}
}
