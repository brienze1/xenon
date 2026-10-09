package adapter

import (
	"errors"

	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/pkg/errs"
)

type Dialog interface {
	PickFolder(title string) (string, error)
	PickFile(title string) (string, error)
	PickSaveFile(title, suggestedName string) (string, error)
}

type wailsDialog struct{}

func NewDialog() Dialog {
	return &wailsDialog{}
}

func (w *wailsDialog) PickFolder(title string) (string, error) {
	app, err := runningApp()
	if err != nil {
		return "", err
	}
	path, err := app.Dialog.OpenFile().
		CanChooseFiles(false).
		CanChooseDirectories(true).
		CanCreateDirectories(true).
		SetTitle(title).
		PromptForSingleSelection()
	if err != nil {
		return "", errs.InternalServerError(err, "DLG-01500", "Folder picker failed")
	}
	return path, nil
}

func (w *wailsDialog) PickFile(title string) (string, error) {
	app, err := runningApp()
	if err != nil {
		return "", err
	}
	path, err := app.Dialog.OpenFile().
		CanChooseFiles(true).
		CanChooseDirectories(false).
		SetTitle(title).
		PromptForSingleSelection()
	if err != nil {
		return "", errs.InternalServerError(err, "DLG-02500", "File picker failed")
	}
	return path, nil
}

func (w *wailsDialog) PickSaveFile(title, suggestedName string) (string, error) {
	app, err := runningApp()
	if err != nil {
		return "", err
	}
	path, err := app.Dialog.SaveFile().
		SetMessage(title).
		SetFilename(suggestedName).
		CanCreateDirectories(true).
		PromptForSingleSelection()
	if err != nil {
		return "", errs.InternalServerError(err, "DLG-03500", "Save dialog failed")
	}
	return path, nil
}

func runningApp() (*wails.App, error) {
	app := wails.Get()
	if app == nil {
		return nil, errs.InternalServerError(errors.New("no running application"), "DLG-00500", "Native dialogs are unavailable")
	}
	return app, nil
}
