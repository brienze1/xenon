package external

import "testing"

type fakeEmitter struct{ names []string }

func (f *fakeEmitter) Emit(name string, _ any) { f.names = append(f.names, name) }

type fakeDialog struct{}

func (f *fakeDialog) PickFolder(string) (string, error) { return "/picked", nil }

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
