package adapter

import (
	"errors"

	wails "github.com/wailsapp/wails/v3/pkg/application"

	"github.com/brienze1/xenon/pkg/errs"
)

type Dialog interface {
	PickFolder(title string) (string, error)
}

type wailsDialog struct{}

func NewDialog() Dialog {
	return &wailsDialog{}
}

func (w *wailsDialog) PickFolder(title string) (string, error) {
	app := wails.Get()
	if app == nil {
		return "", errs.InternalServerError(errors.New("no running application"), "DLG-00500", "Native dialogs are unavailable")
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
