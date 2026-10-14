package external

import (
	"context"
	"testing"

	serverAdapter "github.com/brienze1/xenon/internal/infra/server/adapter"
	httpClient "github.com/brienze1/xenon/pkg/http"
)

type fakeEmitter struct{ names []string }

func (f *fakeEmitter) Emit(name string, _ any) { f.names = append(f.names, name) }

type fakeDialog struct{}

func (f *fakeDialog) PickFolder(string) (string, error)           { return "/picked", nil }
func (f *fakeDialog) PickFile(string) (string, error)             { return "/file", nil }
func (f *fakeDialog) PickSaveFile(string, string) (string, error) { return "/saved", nil }
func (f *fakeDialog) PickSources(string) ([]string, error)        { return []string{"/a"}, nil }

func TestEventEmitterFactoryIsASingletonThatCanBeReplaced(t *testing.T) {
	first := GetEventEmitter()
	if first == nil || GetEventEmitter() != first {
		t.Fatal("GetEventEmitter must return one cached instance")
	}
	replacement := &fakeEmitter{}
	SetEventEmitter(replacement)
	GetEventEmitter().Emit("x", nil)
	if len(replacement.names) != 1 {
		t.Fatal("SetEventEmitter must replace the cached instance")
	}
}

func TestDialogFactoryIsASingletonThatCanBeReplaced(t *testing.T) {
	first := GetDialog()
	if first == nil || GetDialog() != first {
		t.Fatal("GetDialog must return one cached instance")
	}
	SetDialog(&fakeDialog{})
	if path, _ := GetDialog().PickFolder("t"); path != "/picked" {
		t.Fatalf("SetDialog must replace the cached instance, got %q", path)
	}
}

type countingClientFactory struct{ calls int }

func (c *countingClientFactory) Build(options httpClient.Options) httpClient.Client {
	c.calls++
	return httpClient.HttpClient(options, nil)
}

func TestHttpClientFactoryIsASingletonThatCanBeReplaced(t *testing.T) {
	first := GetHttpClientFactory()
	if first == nil || GetHttpClientFactory() != first {
		t.Fatal("GetHttpClientFactory must return one cached instance")
	}
	replacement := &countingClientFactory{}
	SetHttpClientFactory(replacement)
	_ = GetHttpClientFactory().Build(httpClient.Options{})
	if replacement.calls != 1 {
		t.Fatal("SetHttpClientFactory must replace the cached instance")
	}
}

type fakeArchive struct{ extracted []string }

func (f *fakeArchive) Extract(archivePath, _ string, _ int64) error {
	f.extracted = append(f.extracted, archivePath)
	return nil
}

func TestArchiveFactoryIsASingletonThatCanBeReplaced(t *testing.T) {
	first := GetArchive()
	if first == nil || GetArchive() != first {
		t.Fatal("GetArchive must return one cached instance")
	}
	replacement := &fakeArchive{}
	SetArchive(replacement)
	if err := GetArchive().Extract("/source.zip", "/target", 1); err != nil || len(replacement.extracted) != 1 {
		t.Fatal("SetArchive must replace the cached instance")
	}
}

type fakeBrowser struct{ opened []string }

func (f *fakeBrowser) OpenURL(url string) error {
	f.opened = append(f.opened, url)
	return nil
}

func TestBrowserFactoryIsASingletonThatCanBeReplaced(t *testing.T) {
	first := GetBrowser()
	if first == nil || GetBrowser() != first {
		t.Fatal("GetBrowser must return one cached instance")
	}
	replacement := &fakeBrowser{}
	SetBrowser(replacement)
	if err := GetBrowser().OpenURL("https://example.test"); err != nil || len(replacement.opened) != 1 {
		t.Fatal("SetBrowser must replace the cached instance")
	}
}

func TestLoopbackFactoryIsASingletonThatCanBeReplaced(t *testing.T) {
	first := GetLoopback()
	if first == nil || GetLoopback() != first {
		t.Fatal("GetLoopback must return one cached instance")
	}
	SetLoopback(first)
	if GetLoopback() != first {
		t.Fatal("SetLoopback must replace the cached instance")
	}
}

type fakeWindowFocuser struct{ focused int }

func (f *fakeWindowFocuser) Focus() { f.focused++ }

func TestWindowFocuserFactoryIsASingletonThatCanBeReplaced(t *testing.T) {
	first := GetWindowFocuser()
	if first == nil || GetWindowFocuser() != first {
		t.Fatal("GetWindowFocuser must return one cached instance")
	}
	replacement := &fakeWindowFocuser{}
	SetWindowFocuser(replacement)
	GetWindowFocuser().Focus()
	if replacement.focused != 1 {
		t.Fatal("SetWindowFocuser must replace the cached instance")
	}
}

type fakeAIAccessListener struct{ clients int }

func (f *fakeAIAccessListener) Listen(int) error                { return nil }
func (f *fakeAIAccessListener) Close()                          {}
func (f *fakeAIAccessListener) Clients() int                    { return f.clients }
func (f *fakeAIAccessListener) Route(serverAdapter.MCPRoutes)   {}
func (f *fakeAIAccessListener) Watch(func(ctx context.Context)) {}

func TestAIAccessListenerFactoryIsASingletonThatCanBeReplaced(t *testing.T) {
	first := GetAIAccessListener()
	if first == nil || GetAIAccessListener() != first {
		t.Fatal("GetAIAccessListener must return one cached instance")
	}
	SetAIAccessListener(&fakeAIAccessListener{clients: 3})
	if clients := GetAIAccessListener().Clients(); clients != 3 {
		t.Fatalf("SetAIAccessListener must replace the cached instance, got %d clients", clients)
	}
}
