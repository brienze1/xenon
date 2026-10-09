package adapter

import (
	"errors"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

func TestEmitWithoutRunningApplicationIsDropped(t *testing.T) {
	NewEventEmitter().Emit("workspace:tree-changed", nil)
}

func TestPickFolderWithoutRunningApplicationFails(t *testing.T) {
	path, err := NewDialog().PickFolder("Choose a folder")
	var custom errs.BaseError
	if !errors.As(err, &custom) || custom.Code() != "DLG-00500" || path != "" {
		t.Fatalf("PickFolder() = %q, %v; want DLG-00500", path, err)
	}
}

func TestPickFileAndSaveFileWithoutRunningApplicationFail(t *testing.T) {
	for name, pick := range map[string]func() (string, error){
		"file": func() (string, error) { return NewDialog().PickFile("Choose a file") },
		"save": func() (string, error) { return NewDialog().PickSaveFile("Save", "body.json") },
	} {
		path, err := pick()
		var custom errs.BaseError
		if !errors.As(err, &custom) || custom.Code() != "DLG-00500" || path != "" {
			t.Fatalf("%s: = %q, %v; want DLG-00500", name, path, err)
		}
	}
}
