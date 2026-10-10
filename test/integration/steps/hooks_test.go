package steps

import (
	"context"
	"os"
	"path/filepath"

	"github.com/brienze1/xenon/internal/infra/dependency/binding"
	"github.com/brienze1/xenon/internal/infra/dependency/consumer"
)

const scryptLogN = "10"

func (t *testUtils) before() error {
	t.api.Reset()
	t.time.Reset()
	t.events.Reset()
	t.dialog.Reset()
	t.trash.Reset()
	t.watcher.Reset()
	t.keychain.Reset()
	t.clipboard.Reset()
	t.logs.Reset()
	t.random.Reset()
	t.restarts = 0
	t.environment = map[string]string{}
	t.readOnly = nil
	t.response = nil
	t.callError = nil
	t.routeResponse = nil
	t.written = map[string]string{}

	tempRoot, err := os.MkdirTemp("", "xenon-bdd-")
	if err != nil {
		return err
	}
	t.tempRoot, err = filepath.EvalSymlinks(tempRoot)
	if err != nil {
		return err
	}
	t.workspaceRoot = filepath.Join(t.tempRoot, "workspaces")
	t.dataDir = filepath.Join(t.tempRoot, "app-data")
	if err := os.MkdirAll(t.workspaceRoot, 0o755); err != nil {
		return err
	}
	resources, err := filepath.Abs(resourcePath(""))
	if err != nil {
		return err
	}
	t.vars = map[string]string{
		"temp_root":      t.tempRoot,
		"workspace_root": t.workspaceRoot,
		"app_data":       t.dataDir,
		"resources":      resources,
	}
	if err := os.Setenv("XENON_SCRYPT_LOG_N", scryptLogN); err != nil {
		return err
	}
	return os.Setenv("XENON_DATA_DIR", t.dataDir)
}

func (t *testUtils) after() {
	for _, path := range t.readOnly {
		_ = os.Chmod(path, 0o755)
	}
	for name, previous := range t.environment {
		_ = os.Setenv(name, previous)
	}
	_ = os.RemoveAll(t.tempRoot)
}

func (t *testUtils) startApp() {
	consumerOnce.Do(func() {
		consumer.GetFileChangeConsumer().Register(context.Background())
	})
	t.bindings = []any{
		binding.GetWorkspaceService(),
		binding.GetTreeService(),
		binding.GetDialogService(),
		binding.GetRequestService(),
		binding.GetUIStateService(),
		binding.GetHistoryService(),
		binding.GetWindowService(),
		binding.GetResponseBodyRoute(),
		binding.GetEnvironmentService(),
		binding.GetVariableService(),
		binding.GetVaultService(),
		binding.GetImportService(),
		binding.GetExampleService(),
		binding.GetSettingsService(),
		binding.GetConsoleService(),
	}
}
