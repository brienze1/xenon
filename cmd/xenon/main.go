package main

import (
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon"
)

func main() {
	app := application.New(application.Options{
		Name:        "Xenon",
		Description: "Local-first desktop API client",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(xenon.Assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:     "Xenon",
		Width:     1280,
		Height:    800,
		MinWidth:  960,
		MinHeight: 600,
		Mac: application.MacWindow{
			InvisibleTitleBarHeight: 40,
			TitleBar:                application.MacTitleBarHiddenInset,
		},
		BackgroundColour: application.NewRGB(10, 10, 12),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
