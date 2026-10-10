package steps

import (
	"context"
	"flag"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"

	"github.com/brienze1/xenon/internal/infra/dependency/binding"
	"github.com/brienze1/xenon/internal/infra/dependency/consumer"
	"github.com/brienze1/xenon/internal/infra/dependency/external"
	httpClient "github.com/brienze1/xenon/pkg/http"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
	"github.com/brienze1/xenon/test/integration/mock"
)

const scryptLogN = "10"

var tags string

func init() {
	flag.StringVar(&tags, "scenarios", "", "tags to run")
}

type testUtils struct {
	bindings      []any
	api           *mock.ApiMock
	time          *mock.Time
	events        *mock.EventEmitter
	dialog        *mock.Dialog
	trash         *mock.Trash
	watcher       *mock.FileWatcher
	keychain      *mock.Keychain
	clipboard     *mock.Clipboard
	logs          *mock.LogOutput
	random        *mock.RandomSource
	vars          map[string]string
	environment   map[string]string
	readOnly      []string
	written       map[string]string
	tempRoot      string
	workspaceRoot string
	dataDir       string
	routeResponse *httptest.ResponseRecorder
	response      []byte
	callError     []byte
	elapsed       time.Duration
	restarts      int
}

var (
	suite        *testUtils
	suiteOnce    sync.Once
	consumerOnce sync.Once
)

func TestFeatures(t *testing.T) {
	flag.Parse()

	runner := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"../features"},
			Tags:     tags,
			Strict:   true,
			TestingT: t,
		},
	}

	if runner.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	suiteOnce.Do(func() {
		suite = &testUtils{
			api:       mock.NewApiServer(),
			time:      mock.NewTime(),
			events:    mock.NewEventEmitter(),
			dialog:    mock.NewDialog(),
			trash:     mock.NewTrash(),
			watcher:   mock.NewFileWatcher(),
			keychain:  mock.NewKeychain(),
			clipboard: mock.NewClipboard(),
			logs:      mock.NewLogOutput(),
			random:    mock.NewRandomSource(),
		}
		if os.Getenv("LOG_LEVEL") == "" {
			_ = os.Setenv("LOG_LEVEL", "off")
		}
		timeutils.SetCurrentTimeProvider(suite.time.Now)
		external.SetEventEmitter(suite.events)
		external.SetDialog(suite.dialog)
		external.SetTrash(suite.trash)
		external.SetFileWatcher(suite.watcher)
		external.SetKeychain(suite.keychain)
		external.SetClipboard(suite.clipboard)
		external.SetRandomSource(suite.random)
		external.SetHttpClientFactory(httpClient.HttpClientFactory(logger.LoggerWrapper()))
	})
	test := suite

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		if err := test.before(); err != nil {
			return ctx, err
		}
		test.startApp()
		return ctx, nil
	})
	ctx.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		test.after()
		return ctx, err
	})

	test.registerGivenWorkspaceSteps(ctx)
	test.registerGivenApiSteps(ctx)
	test.registerGivenDialogSteps(ctx)
	test.registerWhenBindingSteps(ctx)
	test.registerWhenRouteSteps(ctx)
	test.registerThenRouteSteps(ctx)
	test.registerThenResponseSteps(ctx)
	test.registerThenFileSteps(ctx)
	test.registerThenEventSteps(ctx)
	test.registerThenApiSteps(ctx)
	test.registerThenSecretSteps(ctx)
	test.registerTrashSteps(ctx)
	test.registerFileWatcherSteps(ctx)
	test.registerGivenVaultSteps(ctx)
}

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
	t.vars = map[string]string{
		"temp_root":      t.tempRoot,
		"workspace_root": t.workspaceRoot,
		"app_data":       t.dataDir,
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
	logger.SetOutput(os.Stdout)
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
	}
}
