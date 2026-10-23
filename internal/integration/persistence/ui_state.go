package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"sync"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence/model"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/fileutils"
	"github.com/brienze1/xenon/pkg/logger"
)

const (
	workspacesDataDir = "workspaces"
	uiStateFile       = "state.json"
	digestStateKey    = "digest_state"
	allowlistKey      = "cookie_allowlist"
	runnersKey        = "runners"
)

type uiStatePersistence struct {
	mu sync.Mutex
}

func NewUIStatePersistence() adapter.UIStatePersistence {
	return &uiStatePersistence{}
}

func (u *uiStatePersistence) Find(ctx context.Context, localId string) (*entity.UIState, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	data, err := os.ReadFile(uiStatePath(localId))
	if err != nil {
		return nil, errs.FileError(err, "UIS-02500")
	}
	var stored model.UIState
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, errs.UnprocessableEntityError("state.json is not valid JSON", "UIS-06422")
	}
	state := stored.ToEntity()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "tabs": len(state.Tabs)})
	return &state, nil
}

func (u *uiStatePersistence) Save(ctx context.Context, localId string, state entity.UIState) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "tabs": len(state.Tabs)})

	u.mu.Lock()
	defer u.mu.Unlock()
	fields := map[string]json.RawMessage{}
	data, _ := json.Marshal(model.NewUIState(state))
	_ = json.Unmarshal(data, &fields)
	if err := u.patch(localId, fields, "UIS-03500", "UIS-04500"); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "tabs": len(state.Tabs)})
	return nil
}

func (u *uiStatePersistence) Delete(ctx context.Context, localId string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	if err := os.RemoveAll(appDataPath(workspacesDataDir, localId)); err != nil {
		return errs.FileError(err, "UIS-05500")
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId})
	return nil
}

func (u *uiStatePersistence) FindDigestState(ctx context.Context, localId, owner string) (*entity.DigestState, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "owner": owner})

	u.mu.Lock()
	states, err := u.digestStates(localId)
	u.mu.Unlock()
	if err != nil {
		return nil, err
	}
	stored, found := states[owner]
	if !found {
		logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "found": false})
		return nil, nil
	}
	state := stored.ToEntity()

	logger.Debug(ctx, "Finished", state)
	return &state, nil
}

func (u *uiStatePersistence) SaveDigestState(ctx context.Context, localId, owner string, state entity.DigestState) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "owner": owner, "state": state})

	u.mu.Lock()
	defer u.mu.Unlock()
	states, err := u.digestStates(localId)
	if err != nil {
		return err
	}
	states[owner] = model.NewDigestState(state)
	data, _ := json.Marshal(states)
	if err := u.patch(localId, map[string]json.RawMessage{digestStateKey: data}, "UIS-07500", "UIS-08500"); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "owner": owner, "state": state})
	return nil
}

func (u *uiStatePersistence) FindCookieAllowlist(ctx context.Context, localId string) (*entity.CookieAllowlist, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	u.mu.Lock()
	fields, err := readStateFields(localId)
	u.mu.Unlock()
	if err != nil {
		return nil, err
	}
	domains := []string{}
	_ = json.Unmarshal(fields[allowlistKey], &domains)
	allowlist := entity.NewCookieAllowlist(domains)

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "result": allowlist})
	return &allowlist, nil
}

func (u *uiStatePersistence) SaveCookieAllowlist(ctx context.Context, localId string, allowlist entity.CookieAllowlist) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "allowlist": allowlist})

	u.mu.Lock()
	defer u.mu.Unlock()
	data, _ := json.Marshal(allowlist.Domains)
	if err := u.patch(localId, map[string]json.RawMessage{allowlistKey: data}, "UIS-10500", "UIS-11500"); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "allowlist": allowlist})
	return nil
}

func (u *uiStatePersistence) FindRunnerPreferences(ctx context.Context, localId string) (map[string]entity.RunnerPreferences, error) {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId})

	u.mu.Lock()
	runners, err := u.runners(localId)
	u.mu.Unlock()
	if err != nil {
		return nil, err
	}
	preferences := runners.ToEntity()

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "targets": len(preferences)})
	return preferences, nil
}

func (u *uiStatePersistence) SaveRunnerPreferences(ctx context.Context, localId, itemId string, preferences entity.RunnerPreferences) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "itemId": itemId})

	u.mu.Lock()
	defer u.mu.Unlock()
	runners, err := u.runners(localId)
	if err != nil {
		return err
	}
	runners[itemId] = model.NewRunConfig(preferences.Config)
	if err := u.patchRunners(localId, runners); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "itemId": itemId})
	return nil
}

func (u *uiStatePersistence) DeleteRunnerPreferences(ctx context.Context, localId string, itemIds []string) error {
	logger.Debug(ctx, "Started", map[string]any{"localId": localId, "itemIds": itemIds})

	u.mu.Lock()
	defer u.mu.Unlock()
	runners, err := u.runners(localId)
	if err != nil {
		return err
	}
	for _, itemId := range itemIds {
		delete(runners, itemId)
	}
	if err := u.patchRunners(localId, runners); err != nil {
		return err
	}

	logger.Debug(ctx, "Finished", map[string]any{"localId": localId, "itemIds": itemIds})
	return nil
}

func (u *uiStatePersistence) runners(localId string) (model.RunnerPreferences, error) {
	fields, err := readStateFields(localId)
	if err != nil {
		return nil, err
	}
	runners := model.RunnerPreferences{}
	_ = json.Unmarshal(fields[runnersKey], &runners)
	return runners, nil
}

func (u *uiStatePersistence) patchRunners(localId string, runners model.RunnerPreferences) error {
	data, _ := json.Marshal(runners)
	return u.patch(localId, map[string]json.RawMessage{runnersKey: data}, "UIS-16500", "UIS-17500")
}

func (u *uiStatePersistence) digestStates(localId string) (map[string]model.DigestState, error) {
	fields, err := readStateFields(localId)
	if err != nil {
		return nil, err
	}
	states := map[string]model.DigestState{}
	_ = json.Unmarshal(fields[digestStateKey], &states)
	return states, nil
}

func (u *uiStatePersistence) patch(localId string, updates map[string]json.RawMessage, dirCode, writeCode string) error {
	fields, err := readStateFields(localId)
	if err != nil {
		return err
	}
	for key, value := range updates {
		fields[key] = value
	}
	if err := os.MkdirAll(appDataPath(workspacesDataDir, localId), appDataDirMode); err != nil {
		return errs.FileError(err, dirCode)
	}
	data, _ := json.MarshalIndent(fields, "", "  ")
	if err := fileutils.WriteFileAtomic(uiStatePath(localId), append(data, '\n'), appDataFileMode); err != nil {
		return errs.FileError(err, writeCode)
	}
	return nil
}

func readStateFields(localId string) (map[string]json.RawMessage, error) {
	fields := map[string]json.RawMessage{}
	data, err := os.ReadFile(uiStatePath(localId))
	if errors.Is(err, fs.ErrNotExist) {
		return fields, nil
	}
	if err != nil {
		return nil, errs.FileError(err, "UIS-09500")
	}
	if json.Unmarshal(data, &fields) != nil {
		return map[string]json.RawMessage{}, nil
	}
	return fields, nil
}

func uiStatePath(localId string) string {
	return appDataPath(workspacesDataDir, localId, uiStateFile)
}
