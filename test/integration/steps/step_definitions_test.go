package steps

import (
	"context"
	"flag"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/cucumber/godog"

	application "github.com/brienze1/xenon/internal/infra"
	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/pkg/timeutils"
	"github.com/brienze1/xenon/test/integration/mock"
)

var tags string

func init() {
	flag.StringVar(&tags, "scenarios", "", "tags to run")
}

type testUtils struct {
	bindings      *application.Bindings
	api           *mock.ApiMock
	time          *mock.Time
	events        *mock.EventEmitter
	dialog        *mock.Dialog
	vars          map[string]string
	tempRoot      string
	workspaceRoot string
	dataDir       string
	routeResponse *httptest.ResponseRecorder
	response      []byte
	callError     []byte
}

var (
	suite     *testUtils
	suiteOnce sync.Once
	startOnce sync.Once
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
			api:    mock.NewApiServer(),
			time:   mock.NewTime(),
			events: mock.NewEventEmitter(),
			dialog: mock.NewDialog(),
		}
		if os.Getenv("LOG_LEVEL") == "" {
			_ = os.Setenv("LOG_LEVEL", "off")
		}
		timeutils.SetCurrentTimeProvider(suite.time.Now)
		external.SetEventEmitter(suite.events)
		external.SetDialog(suite.dialog)
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
		_ = os.RemoveAll(test.tempRoot)
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
}

func (t *testUtils) before() error {
	t.api.Start()
	t.time.Reset()
	t.events.Reset()
	t.dialog.Reset()
	t.response = nil
	t.callError = nil
	t.routeResponse = nil

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
	return os.Setenv("XENON_DATA_DIR", t.dataDir)
}

func (t *testUtils) startApp() {
	startOnce.Do(func() {
		t.bindings = application.NewBindings()
	})
}
