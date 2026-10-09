package adapter

import (
	"errors"

	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type wailsDialog struct{}

func NewDialog() adapter.Dialog {
	return &wailsDialog{}
}

func (w *wailsDialog) PickFolder(title string) (string, error) {
	app, err := runningApp()
	if err != nil {
		return "", err
	}
	return app.Dialog.OpenFile().
		CanChooseFiles(false).
		CanChooseDirectories(true).
		CanCreateDirectories(true).
		SetTitle(title).
		PromptForSingleSelection()
}

func (w *wailsDialog) PickFile(title string) (string, error) {
	app, err := runningApp()
	if err != nil {
		return "", err
	}
	return app.Dialog.OpenFile().
		CanChooseFiles(true).
		CanChooseDirectories(false).
		SetTitle(title).
		PromptForSingleSelection()
}

func (w *wailsDialog) PickSaveFile(title, suggestedName string) (string, error) {
	app, err := runningApp()
	if err != nil {
		return "", err
	}
	return app.Dialog.SaveFile().
		SetMessage(title).
		SetFilename(suggestedName).
		CanCreateDirectories(true).
		PromptForSingleSelection()
}

func runningApp() (*wails.App, error) {
	app := wails.Get()
	if app == nil {
		return nil, errors.New("no running application")
	}
	return app, nil
}
