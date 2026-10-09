package service

import (
	"context"
	"errors"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type fakeWindowState struct {
	state *entity.WindowState
	err   error
	saved *entity.WindowState
}

func (f *fakeWindowState) FindWindow(context.Context) (*entity.WindowState, error) {
	return f.state, f.err
}

func (f *fakeWindowState) SaveWindow(_ context.Context, state entity.WindowState) error {
	f.saved = &state
	return nil
}

var primaryScreen = []entity.Screen{{Width: 4000, Height: 3000, Primary: true}}

func TestRestoreWindowReturnsTheSavedState(t *testing.T) {
	saved := entity.WindowState{X: 10, Y: 20, Width: 900, Height: 700}
	store := &fakeWindowState{state: &saved}

	state, err := NewManageWindowStateService(store, store).Restore(context.Background(), primaryScreen)

	if err != nil || *state != saved.FittedTo(primaryScreen) {
		t.Fatalf("Restore() = %+v, %v; want %+v", state, err, saved)
	}
}

func TestRestoreWindowFallsBackToTheDefaults(t *testing.T) {
	for name, failure := range map[string]error{
		"not found":     errs.NotFoundError("missing", "TST-00404"),
		"unprocessable": errs.UnprocessableEntityError("invalid", "TST-00422"),
	} {
		t.Run(name, func(t *testing.T) {
			store := &fakeWindowState{err: failure}

			state, err := NewManageWindowStateService(store, store).Restore(context.Background(), primaryScreen)

			if err != nil || *state != entity.NewWindowState().FittedTo(primaryScreen) {
				t.Fatalf("Restore() = %+v, %v; want the defaults", state, err)
			}
		})
	}
}

func TestRestoreWindowPassesOtherErrorsThrough(t *testing.T) {
	failure := errs.InternalServerError(errors.New("boom"), "TST-00500", "failed")
	store := &fakeWindowState{err: failure}

	state, err := NewManageWindowStateService(store, store).Restore(context.Background(), primaryScreen)

	if state != nil || !errors.Is(err, failure) {
		t.Fatalf("Restore() = %+v, %v; want %v", state, err, failure)
	}
}

func TestSaveWindowUpdatesTheDefaultsWhenNothingIsSaved(t *testing.T) {
	store := &fakeWindowState{err: errs.NotFoundError("missing", "TST-00404")}
	next := entity.WindowState{X: 5, Y: 6, Width: 1000, Height: 800}

	if err := NewManageWindowStateService(store, store).Save(context.Background(), next); err != nil {
		t.Fatalf("Save() = %v", err)
	}
	if want := entity.NewWindowState().Updated(next); store.saved == nil || *store.saved != want {
		t.Fatalf("saved = %+v, want %+v", store.saved, want)
	}
}

func TestSaveWindowPassesOtherErrorsThrough(t *testing.T) {
	failure := errs.InternalServerError(errors.New("boom"), "TST-00500", "failed")
	store := &fakeWindowState{err: failure}

	err := NewManageWindowStateService(store, store).Save(context.Background(), entity.WindowState{Width: 1000, Height: 800})

	if !errors.Is(err, failure) || store.saved != nil {
		t.Fatalf("Save() = %v, saved = %+v", err, store.saved)
	}
}
