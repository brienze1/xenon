package entity

import (
	"errors"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

func TestImportFolderListingStartsEmptyWithTheSkippedEntriesItWasGiven(t *testing.T) {
	listing := NewImportFolderListing("exports", []ImportSkippedEntry{NewSkippedFolderEntry("sub")})
	if !listing.IsEmpty() || listing.Inputs == nil {
		t.Errorf("inputs = %+v, want an empty non-nil slice", listing.Inputs)
	}
	if !listing.HasSkipped() || len(listing.Skipped) != 1 {
		t.Errorf("skipped = %+v", listing.Skipped)
	}
}

func TestImportFolderListingAddKeepsTheInputsInOrder(t *testing.T) {
	listing := NewImportFolderListing("exports", nil)
	listing.Add(ImportInput{Path: "a.json"})
	listing.Add(ImportInput{Path: "b.json"})
	if listing.IsEmpty() || len(listing.Inputs) != 2 || listing.Inputs[0].Path != "a.json" || listing.Inputs[1].Path != "b.json" {
		t.Errorf("inputs = %+v", listing.Inputs)
	}
	if listing.HasSkipped() {
		t.Error("a listing with no skipped entries has nothing skipped")
	}
}

func TestImportFolderListingSkipRecordsTheEntryWithItsReason(t *testing.T) {
	listing := NewImportFolderListing("exports", []ImportSkippedEntry{NewSkippedFolderEntry("sub")})
	listing.Skip("notes.txt", errors.New("notes.txt is not a Xenon, Postman, Bruno or cURL export"))
	want := []ImportSkippedEntry{{Name: "sub", Reason: "folder"}, {Name: "notes.txt", Reason: "notes.txt is not a Xenon, Postman, Bruno or cURL export"}}
	if len(listing.Skipped) != len(want) || listing.Skipped[0] != want[0] || listing.Skipped[1] != want[1] {
		t.Errorf("skipped = %+v, want %+v", listing.Skipped, want)
	}
	if !listing.IsEmpty() {
		t.Error("skipping an entry adds no input")
	}
}

func TestImportFolderListingSkippedWarningNamesTheFolderAndEachEntry(t *testing.T) {
	listing := NewImportFolderListing("exports", []ImportSkippedEntry{NewSkippedFolderEntry("sub")})
	listing.Skip("notes.txt", errors.New("notes.txt is not a Xenon, Postman, Bruno or cURL export"))
	warning := listing.SkippedWarning()
	want := "Skipped in exports: notes.txt (not a Xenon, Postman, Bruno or cURL export), sub (folder)"
	if warning.Code != warningcode.SkippedPart || warning.Message != want {
		t.Errorf("warning = %+v", warning)
	}
}
