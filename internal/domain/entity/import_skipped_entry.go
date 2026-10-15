package entity

import (
	"errors"
	"strings"

	"github.com/brienze1/xenon/pkg/errs"
)

const skippedFolderReason = "folder"

type ImportSkippedEntry struct {
	Name   string
	Reason string
}

func NewImportSkippedEntry(name string, err error) ImportSkippedEntry {
	var custom errs.BaseError
	if errors.As(err, &custom) {
		return ImportSkippedEntry{Name: name, Reason: custom.Description()}
	}
	return ImportSkippedEntry{Name: name, Reason: err.Error()}
}

func NewSkippedFolderEntry(name string) ImportSkippedEntry {
	return ImportSkippedEntry{Name: name, Reason: skippedFolderReason}
}

func (i ImportSkippedEntry) ShortReason() string {
	predicate, named := strings.CutPrefix(i.Reason, i.Name+" ")
	if !named {
		return i.Reason
	}
	return strings.TrimPrefix(predicate, "is ")
}

func compareSkippedEntries(a, b ImportSkippedEntry) int {
	return strings.Compare(a.Name, b.Name)
}
