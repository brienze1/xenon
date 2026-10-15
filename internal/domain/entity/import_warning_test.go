package entity

import (
	"errors"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
	"github.com/brienze1/xenon/internal/domain/enums/warningseverity"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestNewSkippedEntriesWarningNamesEachEntryWithItsReason(t *testing.T) {
	entries := []ImportSkippedEntry{
		{Name: "sub", Reason: "folder"},
		{Name: "notes.txt", Reason: "notes.txt is not a Xenon, Postman, Bruno or cURL export"},
		{Name: "old.json", Reason: "Postman v1 collections are not supported; export the collection as v2.1"},
	}
	warning := NewSkippedEntriesWarning("with-extras", entries)
	want := "Skipped in with-extras: notes.txt (not a Xenon, Postman, Bruno or cURL export), old.json (Postman v1 collections are not supported; export the collection as v2.1), sub (folder)"
	if warning.Code != warningcode.SkippedPart || warning.Severity != warningseverity.Warning || warning.Message != want {
		t.Errorf("warning = %+v", warning)
	}
	if entries[0].Name != "sub" {
		t.Error("the entries passed in keep their order")
	}
}

func TestNewImportSkippedEntryKeepsTheErrorDescription(t *testing.T) {
	cases := map[string]struct {
		entry ImportSkippedEntry
		want  ImportSkippedEntry
	}{
		"custom error": {NewImportSkippedEntry("a.json", errs.BadRequestError("a.json is not valid JSON", "IMP-01400")), ImportSkippedEntry{Name: "a.json", Reason: "a.json is not valid JSON"}},
		"plain error":  {NewImportSkippedEntry("b.zip", errors.New("broken")), ImportSkippedEntry{Name: "b.zip", Reason: "broken"}},
		"folder":       {NewSkippedFolderEntry("sub"), ImportSkippedEntry{Name: "sub", Reason: "folder"}},
	}
	for name, tc := range cases {
		if tc.entry != tc.want {
			t.Errorf("%s: got %+v, want %+v", name, tc.entry, tc.want)
		}
	}
}
