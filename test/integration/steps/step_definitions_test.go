package steps

import (
	"context"
	"flag"
	"net"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/cucumber/godog"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	httpClient "github.com/brienze1/xenon/pkg/http"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
	"github.com/brienze1/xenon/test/integration/mock"
)

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
	oauth         *mock.OAuthProvider
	browser       *mock.BrowserOpener
	digest        *mock.DigestServer
	fileManager   *mock.FileManager
	git           *mock.GitServers
	sshAgent      *mock.SshAgent
	window        *mock.WindowFocuser
	process       *mock.AppProcess
	mcpSessions   map[string]*mcp.ClientSession
	bridge        *bridgeRun
	sshKeys       map[string]*mock.SshKey
	busyPorts     []net.Listener
	mcpClientName string
	vars          map[string]string
	environment   map[string]string
	readOnly      []string
	written       map[string]string
	tempRoot      string
	workspaceRoot string
	dataDir       string
	routeResponse *httptest.ResponseRecorder
	background    *backgroundCall
	backgrounds   map[string]*backgroundCall
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
			api:         mock.NewApiServer(),
			time:        mock.NewTime(),
			events:      mock.NewEventEmitter(),
			dialog:      mock.NewDialog(),
			trash:       mock.NewTrash(),
			watcher:     mock.NewFileWatcher(),
			keychain:    mock.NewKeychain(),
			clipboard:   mock.NewClipboard(),
			logs:        mock.NewLogOutput(),
			random:      mock.NewRandomSource(),
			oauth:       mock.NewOAuthProvider(),
			browser:     mock.NewBrowserOpener(),
			digest:      mock.NewDigestServer(),
			fileManager: mock.NewFileManager(),
			git:         mock.NewGitServers(),
			sshAgent:    mock.NewSshAgent(),
			window:      mock.NewWindowFocuser(),
			process:     mock.NewAppProcess(),
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
		external.SetBrowser(suite.browser)
		external.SetFileManager(suite.fileManager)
		external.SetSshAgent(suite.sshAgent)
		external.SetTimers(suite.time)
		external.SetWindowFocuser(suite.window)
		external.SetAppProcess(suite.process)
		external.SetHttpClientFactory(httpClient.HttpClientFactory(logger.LoggerWrapper()))
		logger.SetOutput(suite.logs)
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
	test.registerGivenAppDataSteps(ctx)
	test.registerGivenTimeSteps(ctx)
	test.registerGivenVariableSteps(ctx)
	test.registerGivenApiSteps(ctx)
	test.registerGivenDialogSteps(ctx)
	test.registerWhenBindingSteps(ctx)
	test.registerWhenProgramSteps(ctx)
	test.registerWhenRouteSteps(ctx)
	test.registerThenRouteSteps(ctx)
	test.registerThenResponseSteps(ctx)
	test.registerThenFileSteps(ctx)
	test.registerThenExportFileSteps(ctx)
	test.registerThenFileCompareSteps(ctx)
	test.registerFileManagerSteps(ctx)
	test.registerGivenResourceSteps(ctx)
	test.registerWhenNewmanSteps(ctx)
	test.registerThenApiCompareSteps(ctx)
	test.registerGivenEventSteps(ctx)
	test.registerThenEventSteps(ctx)
	test.registerThenApiSteps(ctx)
	test.registerThenApiBodySteps(ctx)
	test.registerThenWorkspaceFileSteps(ctx)
	test.registerThenAppDataSteps(ctx)
	test.registerThenKeychainSteps(ctx)
	test.registerThenClipboardSteps(ctx)
	test.registerThenLogSteps(ctx)
	test.registerThenResponseContentSteps(ctx)
	test.registerTrashSteps(ctx)
	test.registerFileWatcherSteps(ctx)
	test.registerGivenKeychainSteps(ctx)
	test.registerGivenAppSteps(ctx)
	test.registerAppProcessSteps(ctx)
	test.registerGivenEnvSteps(ctx)
	test.registerGivenLogSteps(ctx)
	test.registerGivenOAuthSteps(ctx)
	test.registerThenOAuthSteps(ctx)
	test.registerGivenDigestSteps(ctx)
	test.registerThenDigestSteps(ctx)
	test.registerGivenGitSteps(ctx)
	test.registerThenGitSteps(ctx)
	test.registerWhenMcpSteps(ctx)
	test.registerWhenMcpHttpSteps(ctx)
	test.registerWhenMcpBridgeSteps(ctx)
	test.registerThenMcpSteps(ctx)
	test.registerThenMcpBridgeSteps(ctx)
}
