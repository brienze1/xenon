package application

import (
	"context"
	"log"
	"path/filepath"

	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/infra/server"
	"github.com/brienze1/xenon/internal/infra/server/binding"
	_ "github.com/brienze1/xenon/internal/infra/server/event"
	"github.com/brienze1/xenon/pkg/logger"
)

type Bindings struct {
	Dialog *binding.DialogService
}

func NewBindings() *Bindings {
	return &Bindings{
		Dialog: binding.NewDialogService(),
	}
}

func Start() {
	ctx := context.Background()
	sink, err := logger.SetFileSink(filepath.Join(properties.Properties().Application.DataDir, "logs", "xenon.log"))
	if err != nil {
		logger.Warn(ctx, err, "Log file unavailable, logging to stdout")
	} else {
		defer func() { _ = sink.Close() }()
	}
	logger.Info(ctx, "Starting application", map[string]any{"version": properties.Properties().Application.Version})

	bindings := NewBindings()
	var window *wails.WebviewWindow
	app := wails.New(wails.Options{
		Name:        "Xenon",
		Description: "Local-first desktop API client",
		Services: []wails.Service{
			wails.NewService(bindings.Dialog),
		},
		MarshalError: server.MarshalError,
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
		MinWidth:  960,
		MinHeight: 600,
		Mac: wails.MacWindow{
			InvisibleTitleBarHeight: 40,
			TitleBar:                wails.MacTitleBarHiddenInset,
		},
		BackgroundColour: wails.NewRGB(10, 10, 12),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
