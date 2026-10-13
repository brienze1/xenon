package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
)

type fakeAppSettings struct {
	settings entity.AppSettings
	saved    *entity.AppSettings
	err      error
}

func (f *fakeAppSettings) FindAppSettings(context.Context) (*entity.AppSettings, error) {
	settings := f.settings
	return &settings, f.err
}

func (f *fakeAppSettings) SaveAppSettings(_ context.Context, settings entity.AppSettings) error {
	f.saved = &settings
	return nil
}

type fakeServeAIAccess struct {
	applied []entity.AIAccessSettings
	stopped int
}

func (f *fakeServeAIAccess) Apply(_ context.Context, settings entity.AIAccessSettings) entity.AIAccessStatus {
	f.applied = append(f.applied, settings)
	return entity.RunningAIAccessStatus(settings.EffectivePort(nil), 0)
}

func (f *fakeServeAIAccess) Status(context.Context) entity.AIAccessStatus {
	return entity.RunningAIAccessStatus(entity.DefaultAIPort, 2)
}

func (f *fakeServeAIAccess) Stop(context.Context) entity.AIAccessStatus {
	f.stopped++
	return entity.OffAIAccessStatus(entity.DefaultAIPort)
}

type fakeVault struct {
	vault *entity.Vault
	err   error
}

func (f fakeVault) Find(context.Context, string) (*entity.Vault, error) {
	return f.vault, f.err
}

type fakeKnownSecrets struct {
	known entity.KnownSecrets
}

func (f *fakeKnownSecrets) Register(_ context.Context, values ...string) entity.KnownSecrets {
	f.known.Add(values...)
	return f.known.With()
}

type fakeHistoryStore struct {
	saved   []entity.HistoryEntry
	caps    entity.HistoryCaps
	cleared enums.HistoryFilter
	err     error
}

func (f *fakeHistoryStore) Save(_ context.Context, _ string, entry entity.HistoryEntry, caps entity.HistoryCaps) error {
	f.saved, f.caps = append(f.saved, entry), caps
	return f.err
}

func (f *fakeHistoryStore) Find(_ context.Context, _ string, caps entity.HistoryCaps) (*entity.History, error) {
	f.caps = caps
	history := entity.NewHistory(f.saved, caps)
	return &history, f.err
}

func (f *fakeHistoryStore) FindEntry(_ context.Context, _ string, id string, caps entity.HistoryCaps) (*entity.HistoryEntry, error) {
	f.caps = caps
	for _, entry := range f.saved {
		if entry.HasID(id) {
			return &entry, nil
		}
	}
	return nil, errors.New("missing")
}

func (f *fakeHistoryStore) Clear(_ context.Context, _ string, source enums.HistoryFilter) error {
	f.cleared = source
	return f.err
}

type fakeEvents struct {
	names []string
}

func (f *fakeEvents) AccessStatusChanged(context.Context, entity.AIAccessStatus) error {
	f.names = append(f.names, "access")
	return nil
}
func (f *fakeEvents) ApprovalRequested(context.Context, entity.AIApproval) error {
	f.names = append(f.names, "requested")
	return nil
}
func (f *fakeEvents) ApprovalResolved(context.Context, entity.AIApproval) error {
	f.names = append(f.names, "resolved")
	return nil
}
func (f *fakeEvents) HistoryChanged(_ context.Context, localId string) error {
	f.names = append(f.names, "history:"+localId)
	return nil
}
func (f *fakeEvents) WorkspacesChanged(_ context.Context, active string) error {
	f.names = append(f.names, "workspaces:"+active)
	return nil
}
func (f *fakeEvents) TreeChanged(_ context.Context, localId string) error {
	f.names = append(f.names, "tree:"+localId)
	return nil
}
func (f *fakeEvents) RequestChangedOnDisk(_ context.Context, _ string, change entity.RequestChange) error {
	f.names = append(f.names, "request:"+change.RequestID)
	return nil
}
func (f *fakeEvents) FolderMissing(context.Context, string) error { return nil }
func (f *fakeEvents) CookiesChanged(_ context.Context, localId string) error {
	f.names = append(f.names, "cookies:"+localId)
	return nil
}
func (f *fakeEvents) VariablesChanged(_ context.Context, localId string, _ []entity.ScopeRef, _ string) error {
	f.names = append(f.names, "variables:"+localId)
	return nil
}
func (f *fakeEvents) VaultStatusChanged(context.Context, entity.VaultStatus) error { return nil }

func TestSaveAppSettingsMergesValidatesAndAppliesAIAccess(t *testing.T) {
	store, server := &fakeAppSettings{settings: entity.DefaultAppSettings()}, &fakeServeAIAccess{}
	service := NewManageAppSettingsService(store, store, server)
	timeout, port, low := 2500, 48000, 99

	saved, err := service.Save(context.Background(), entity.AppSettingsPatch{ScriptTimeoutMs: &timeout})
	if err != nil || saved.ScriptTimeoutMs != 2500 || saved.HistoryAICap != entity.DefaultHistoryCap || len(server.applied) != 0 {
		t.Fatalf("Save(timeout) = %+v, %v, applied %d", saved, err, len(server.applied))
	}
	saved, err = service.Save(context.Background(), entity.AppSettingsPatch{AIAccess: &entity.AIAccessPatch{Port: &port}})
	if err != nil || saved.AIAccess.Port != port || len(server.applied) != 1 || server.applied[0].Port != port {
		t.Fatalf("Save(port) = %+v, %v, applied %+v", saved, err, server.applied)
	}
	store.saved = nil
	if _, err := service.Save(context.Background(), entity.AppSettingsPatch{HistoryUserCap: &low}); err == nil || !strings.Contains(err.Error(), "bad request") || store.saved != nil {
		t.Fatalf("an invalid cap must be refused before saving: %v", err)
	}
	store.err = errors.New("unreadable")
	if _, err := service.Save(context.Background(), entity.AppSettingsPatch{}); err == nil {
		t.Fatal("a read failure must be returned")
	}
	if _, err := service.Get(context.Background()); err == nil {
		t.Fatal("a read failure must be returned by Get")
	}
}

func TestManageAIAccessStartsFromTheSavedSettings(t *testing.T) {
	store, server := &fakeAppSettings{settings: entity.DefaultAppSettings()}, &fakeServeAIAccess{}
	service := NewManageAIAccessService(store, server)
	started, err := service.Start(context.Background())
	if err != nil || !started.IsRunning() || len(server.applied) != 1 || !server.applied[0].Enabled {
		t.Fatalf("Start() = %+v, %v", started, err)
	}
	if status := service.Status(context.Background()); status.Clients != 2 {
		t.Errorf("Status() = %+v", status)
	}
	if stopped := service.Stop(context.Background()); stopped.IsRunning() || server.stopped != 1 {
		t.Errorf("Stop() = %+v", stopped)
	}
	store.err = errors.New("unreadable")
	if _, err := service.Start(context.Background()); err == nil {
		t.Error("a read failure must be returned")
	}
}

func TestGuardAIOutputMasksVaultAndSessionSecrets(t *testing.T) {
	vault := entity.Vault{State: "unlocked", Secrets: map[string]map[string]string{"globals": {"token": "vault-secret"}}}
	known := &fakeKnownSecrets{}
	service := NewGuardAIOutputService(fakeVault{vault: &vault}, known)
	service.Register(context.Background(), "session-secret")

	masked := service.Guard(context.Background(), "ws-1", `{"a":"vault-secret","b":"session-secret","c":"plain"}`)
	if masked != `{"a":"•••••","b":"•••••","c":"plain"}` {
		t.Fatalf("Guard() = %s", masked)
	}
	failing := NewGuardAIOutputService(fakeVault{err: errors.New("locked out")}, known)
	if masked := failing.Guard(context.Background(), "ws-1", "vault-secret"); masked != "•••••" {
		t.Errorf("values seen before stay masked when the vault fails: %s", masked)
	}
	if masked := failing.Guard(context.Background(), "", "session-secret"); masked != "•••••" {
		t.Errorf("no workspace: %s", masked)
	}
}

func TestNotifyAIChangePublishesTheMatchingEvents(t *testing.T) {
	events := &fakeEvents{}
	service := NewNotifyAIChangeService(fakeWorkspaceRegistry{}, events, events, events)
	changes := []entity.AIChange{
		{Kind: aichangekind.Tree},
		{Kind: aichangekind.Request, RequestID: "r1"},
		{Kind: aichangekind.Variables},
		{Kind: aichangekind.History},
		{Kind: aichangekind.Cookies},
		{Kind: aichangekind.Workspaces},
		{Kind: aichangekind.Unknown},
	}
	if err := service.Notify(context.Background(), "ws", changes); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(events.names, ","); got != "tree:ws,request:r1,variables:ws,history:ws,cookies:ws,workspaces:" {
		t.Errorf("events = %s", got)
	}
}

func TestRecordAIActivityAppendsARedactedGuardedActionEntry(t *testing.T) {
	vault := entity.Vault{State: "unlocked", Secrets: map[string]map[string]string{"globals": {"token": "vault-secret"}}}
	settings := entity.DefaultAppSettings()
	settings.HistoryAICap = 100
	history, events := &fakeHistoryStore{}, &fakeEvents{}
	service := NewRecordAIActivityService(&fakeAppSettings{settings: settings}, fakeVault{vault: &vault}, &fakeKnownSecrets{}, history, events)
	activity := entity.AIActivity{Tool: "env_edit", Action: "add_var", Outcome: entity.AIActivityError, Args: `{"action":"add_var","value":"typed-secret","note":"vault-secret"}`}

	service.Record(context.Background(), "ws", "Claude Code", activity)
	if len(history.saved) != 1 || history.caps.AI != 100 || strings.Join(events.names, ",") != "history:ws" {
		t.Fatalf("saved %d entries with caps %+v, events %v", len(history.saved), history.caps, events.names)
	}
	recorded := history.saved[0]
	if !recorded.IsAction() || recorded.Client != "Claude Code" || recorded.Activity.Args != `{"action":"add_var","note":"•••••","value":"•••••"}` {
		t.Errorf("recorded = %+v / %s", recorded, recorded.Activity.Args)
	}
	service.Record(context.Background(), "", "Claude Code", activity)
	history.err = errors.New("disk full")
	service.Record(context.Background(), "ws", "Claude Code", activity)
	if len(history.saved) != 2 {
		t.Errorf("no workspace must only log; a failing save must not panic: %d", len(history.saved))
	}
	failing := NewRecordAIActivityService(&fakeAppSettings{err: errors.New("x")}, fakeVault{}, &fakeKnownSecrets{}, history, events)
	failing.Record(context.Background(), "ws", "c", activity)
	noVault := NewRecordAIActivityService(&fakeAppSettings{settings: settings}, fakeVault{err: errors.New("x")}, &fakeKnownSecrets{}, history, events)
	noVault.Record(context.Background(), "ws", "c", activity)
}

func TestManageHistoryFiltersBySourceAndPublishesClears(t *testing.T) {
	settings := entity.DefaultAppSettings()
	history, events := &fakeHistoryStore{}, &fakeEvents{}
	history.saved = []entity.HistoryEntry{{ID: "u", Source: "user"}, entity.NewActionHistoryEntry("a", entity.HistoryEntry{}.SentAt, "c", entity.AIActivity{Tool: "tree"})}
	service := NewManageHistoryService(fakeWorkspaceRegistry{}, &fakeAppSettings{settings: settings}, history, history, events)

	listed, err := service.List(context.Background(), "ws", entity.NewHistoryQuery("", "", 0).FromSource(historyfilter.AI))
	if err != nil || len(listed) != 1 || listed[0].ID != "a" || history.caps != settings.HistoryCaps() {
		t.Fatalf("List(ai) = %v, %v", listed, err)
	}
	if entry, err := service.Get(context.Background(), "ws", "u"); err != nil || entry.ID != "u" {
		t.Fatalf("Get() = %v, %v", entry, err)
	}
	if err := service.Clear(context.Background(), "ws", historyfilter.User); err != nil || history.cleared != historyfilter.User || strings.Join(events.names, ",") != "history:ws" {
		t.Fatalf("Clear(user) = %v, cleared %s, events %v", err, history.cleared, events.names)
	}
	failing := NewManageHistoryService(fakeWorkspaceRegistry{}, &fakeAppSettings{err: errors.New("x")}, history, history, events)
	if _, err := failing.List(context.Background(), "ws", entity.NewHistoryQuery("", "", 0)); err == nil {
		t.Error("List must return a settings failure")
	}
	if _, err := failing.Get(context.Background(), "ws", "u"); err == nil {
		t.Error("Get must return a settings failure")
	}
	history.err = errors.New("x")
	if err := service.Clear(context.Background(), "ws", historyfilter.All); err == nil {
		t.Error("Clear must return a store failure")
	}
}
