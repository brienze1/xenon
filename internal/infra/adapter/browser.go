package adapter

import "github.com/brienze1/xenon/internal/integration/adapter"

type wailsBrowser struct{}

func NewBrowser() adapter.Browser {
	return &wailsBrowser{}
}

func (w *wailsBrowser) OpenURL(url string) error {
	app, err := runningApp()
	if err != nil {
		return err
	}
	return app.Browser.OpenURL(url)
}
