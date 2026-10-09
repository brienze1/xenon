package adapter

import "testing"

func TestEmitWithoutRunningApplicationIsDropped(t *testing.T) {
	NewEventEmitter().Emit("workspace:tree-changed", nil)
}

func TestPickFolderWithoutRunningApplicationFails(t *testing.T) {
	path, err := NewDialog().PickFolder("Choose a folder")
	if err == nil || path != "" {
		t.Fatalf("PickFolder() = %q, %v; want an error", path, err)
	}
}

func TestPickFileAndSaveFileWithoutRunningApplicationFail(t *testing.T) {
	for name, pick := range map[string]func() (string, error){
		"file": func() (string, error) { return NewDialog().PickFile("Choose a file") },
		"save": func() (string, error) { return NewDialog().PickSaveFile("Save", "body.json") },
	} {
		path, err := pick()
		if err == nil || path != "" {
			t.Fatalf("%s: = %q, %v; want an error", name, path, err)
		}
	}
}
