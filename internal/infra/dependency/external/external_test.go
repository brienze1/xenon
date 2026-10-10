package external

import (
	"testing"

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
