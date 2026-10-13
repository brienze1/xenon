package webservice

import (
	"context"
	"errors"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

type fakeFileManager struct {
	failure error
	shown   string
}

func (f *fakeFileManager) ShowInFolder(path string) error {
	f.shown = path
	return f.failure
}

func TestShowInFolder(t *testing.T) {
	manager := &fakeFileManager{}
	if err := NewFileManagerWebservice(manager).ShowInFolder(context.Background(), "/tmp/x.json"); err != nil || manager.shown != "/tmp/x.json" {
		t.Fatalf("err = %v, shown = %q", err, manager.shown)
	}
	manager.failure = errors.New("no file manager")
	err := NewFileManagerWebservice(manager).ShowInFolder(context.Background(), "/tmp/x.json")
	if custom, isCustom := err.(errs.BaseError); !isCustom || custom.Code() != "EXP-09500" {
		t.Fatalf("err = %v", err)
	}
}
