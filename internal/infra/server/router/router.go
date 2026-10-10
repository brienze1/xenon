package router

import (
	wails "github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"

	"github.com/brienze1/xenon"
	"github.com/brienze1/xenon/internal/domain/entity"
	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/infra/server/binding"
	"github.com/brienze1/xenon/internal/infra/server/errorpayload"
	_ "github.com/brienze1/xenon/internal/infra/server/event"
)

type router struct {
	workspace    *binding.WorkspaceService
	tree         *binding.TreeService
	dialog       *binding.DialogService
	request      *binding.RequestService
	uiState      *binding.UIStateService
	history      *binding.HistoryService
	window       *binding.WindowService
	responseBody *binding.ResponseBodyRoute
	environment  *binding.EnvironmentService
	variable     *binding.VariableService
	vault        *binding.VaultService
	importer     *binding.ImportService
	example      *binding.ExampleService
}

func NewRouter(
	workspace *binding.WorkspaceService,
	tree *binding.TreeService,
	dialog *binding.DialogService,
	request *binding.RequestService,
	uiState *binding.UIStateService,
	history *binding.HistoryService,
	window *binding.WindowService,
	responseBody *binding.ResponseBodyRoute,
	environment *binding.EnvironmentService,
	variable *binding.VariableService,
	vault *binding.VaultService,
	importer *binding.ImportService,
	example *binding.ExampleService,
) serverAdapter.Router {
	return &router{
		workspace:    workspace,
		tree:         tree,
		dialog:       dialog,
		request:      request,
		uiState:      uiState,
		history:      history,
		window:       window,
		responseBody: responseBody,
		environment:  environment,
		variable:     variable,
		vault:        vault,
		importer:     importer,
		example:      example,
	}
}

func (r *router) Route() *wails.App {
	var window *wails.WebviewWindow
	app := wails.New(wails.Options{
		Name:         "Xenon",
		Description:  "Local-first desktop API client",
		Services:     r.registerServices(),
		MarshalError: errorpayload.MarshalError,
		Assets: wails.AssetOptions{
			Handler: wails.AssetFileServerFS(xenon.Assets),
		},
		SingleInstance: &wails.SingleInstanceOptions{
			UniqueID: "dev.brienze.xenon",
			OnSecondInstanceLaunch: func(wails.SecondInstanceData) {
				if window != nil {
					window.Restore()
					window.Focus()
				}
			},
		},
		Mac: wails.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	window = app.Window.NewWithOptions(wails.WebviewWindowOptions{
		Title:     "Xenon",
		Width:     1280,
		Height:    800,
		MinWidth:  entity.MinWindowWidth,
		MinHeight: entity.MinWindowHeight,
		Hidden:    true,
		Mac: wails.MacWindow{
			InvisibleTitleBarHeight: 40,
			TitleBar:                wails.MacTitleBarHiddenInset,
		},
		BackgroundColour: wails.NewRGB(10, 10, 12),
		URL:              "/",
		EnableFileDrop:   true,
	})
	r.registerMenuRoutes(app)

	app.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*wails.ApplicationEvent) {
		r.registerWindowRoutes(app, window)
		r.registerDropRoutes(app, window)
	})

	return app
}

func (r *router) registerServices() []wails.Service {
	return []wails.Service{
		r.registerWorkspaceRoutes(),
		r.registerTreeRoutes(),
		r.registerDialogRoutes(),
		r.registerRequestRoutes(),
		r.registerUIStateRoutes(),
		r.registerHistoryRoutes(),
		r.registerResponseBodyRoutes(),
		r.registerEnvironmentRoutes(),
		r.registerVariableRoutes(),
		r.registerVaultRoutes(),
		r.registerImportRoutes(),
		r.registerExampleRoutes(),
	}
}
