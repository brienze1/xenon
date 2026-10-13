package adapter

import "github.com/brienze1/xenon/internal/integration/adapter"

type wailsFileManager struct{}

func NewFileManager() adapter.FileManager {
	return &wailsFileManager{}
}

func (w *wailsFileManager) ShowInFolder(path string) error {
	app, err := runningApp()
	if err != nil {
		return err
	}
	return app.Env.OpenFileManager(path, true)
}
