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
