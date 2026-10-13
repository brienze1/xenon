package mock

import (
	"io"
	"net/http"
	"net/url"
	"sync"
)

type BrowserOpener struct {
	client  *http.Client
	opened  []string
	visited []string
	ignore  bool
	mu      sync.Mutex
}

func NewBrowserOpener() *BrowserOpener {
	return &BrowserOpener{client: &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}}}
}

func (b *BrowserOpener) OpenURL(target string) error {
	b.mu.Lock()
	b.opened = append(b.opened, target)
	ignore := b.ignore
	b.mu.Unlock()
	if !ignore {
		go b.follow(target)
	}
	return nil
}

func (b *BrowserOpener) IgnoreOpenedPages() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ignore = true
}

func (b *BrowserOpener) Opened() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]string{}, b.opened...)
}

func (b *BrowserOpener) Visited() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return append([]string{}, b.visited...)
}

func (b *BrowserOpener) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.opened = nil
	b.visited = nil
	b.ignore = false
}

func (b *BrowserOpener) follow(target string) {
	for hop := 0; hop < 5 && target != ""; hop++ {
		target = b.visit(target)
	}
}

func (b *BrowserOpener) visit(target string) string {
	parsed, err := url.Parse(target)
	if err != nil {
		return ""
	}
	if parsed.Fragment != "" {
		parsed.RawQuery = parsed.Fragment
		parsed.Fragment = ""
	}
	b.mu.Lock()
	b.visited = append(b.visited, parsed.String())
	b.mu.Unlock()
	response, err := b.client.Get(parsed.String())
	if err != nil {
		return ""
	}
	defer func() { _ = response.Body.Close() }()
	_, _ = io.Copy(io.Discard, response.Body)
	location, err := response.Location()
	if err != nil {
		return ""
	}
	return location.String()
}
