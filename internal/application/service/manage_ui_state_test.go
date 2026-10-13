package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type fakeWorkspaceRegistry struct{}

func (fakeWorkspaceRegistry) Find(context.Context) (*entity.AppRegistry, error) {
	registry := entity.NewAppRegistry()
	return &registry, nil
}

func (fakeWorkspaceRegistry) FindWorkspace(context.Context, string) (*entity.RegisteredWorkspace, error) {
	return &entity.RegisteredWorkspace{}, nil
}

type fakeUIState struct {
	state *entity.UIState
	err   error
}

func (f fakeUIState) Find(context.Context, string) (*entity.UIState, error) {
	return f.state, f.err
}

func (f fakeUIState) Save(context.Context, string, entity.UIState) error {
	return nil
}

func TestGetUIStateReturnsTheSavedState(t *testing.T) {
	saved := entity.UIState{Version: 1, ActiveTabID: "tab-1"}
	service := NewManageUIStateService(fakeWorkspaceRegistry{}, fakeUIState{state: &saved}, fakeUIState{})

	state, err := service.Get(context.Background(), "local-1")

	if err != nil || !reflect.DeepEqual(*state, saved) {
		t.Fatalf("Get() = %+v, %v; want %+v", state, err, saved)
	}
}

func TestGetUIStateFallsBackToTheDefaults(t *testing.T) {
	for name, failure := range map[string]error{
		"not found":     errs.NotFoundError("missing", "TST-00404"),
		"unprocessable": errs.UnprocessableEntityError("invalid", "TST-00422"),
	} {
		t.Run(name, func(t *testing.T) {
			service := NewManageUIStateService(fakeWorkspaceRegistry{}, fakeUIState{err: failure}, fakeUIState{})

			state, err := service.Get(context.Background(), "local-1")

			if err != nil || !reflect.DeepEqual(*state, entity.NewUIState()) {
				t.Fatalf("Get() = %+v, %v; want the defaults", state, err)
			}
		})
	}
}

func TestGetUIStatePassesOtherErrorsThrough(t *testing.T) {
	failure := errs.InternalServerError(errors.New("boom"), "TST-00500", "failed")
	service := NewManageUIStateService(fakeWorkspaceRegistry{}, fakeUIState{err: failure}, fakeUIState{})

	state, err := service.Get(context.Background(), "local-1")

	if state != nil || !errors.Is(err, failure) {
		t.Fatalf("Get() = %+v, %v; want %v", state, err, failure)
	}
}
