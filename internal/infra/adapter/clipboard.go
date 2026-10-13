package adapter

import (
	"errors"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type wailsClipboard struct{}

func NewClipboard() adapter.Clipboard {
	return &wailsClipboard{}
}

func (w *wailsClipboard) SetText(text string) error {
	app, err := runningApp()
	if err != nil {
		return err
	}
	if !app.Clipboard.SetText(text) {
		return errors.New("the clipboard rejected the text")
	}
	return nil
}
