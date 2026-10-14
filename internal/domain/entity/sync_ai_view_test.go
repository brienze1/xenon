package entity

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/conflictkind"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
	"github.com/brienze1/xenon/internal/domain/enums/syncstate"
)

func TestRemoteWithoutCredentials(t *testing.T) {
	cases := map[string]string{
		"https://user:token@github.com/acme/api.git": "https://github.com/acme/api.git",
		"https://github.com/acme/api.git":            "https://github.com/acme/api.git",
		"git@github.com:acme/api.git":                "git@github.com:acme/api.git",
		"":                                           "",
		"://bad":                                     "://bad",
	}
	for repository, expected := range cases {
		if remote := (LocalSyncSettings{Repository: repository}).Remote(); remote != expected {
			t.Errorf("Remote(%q) = %q, want %q", repository, remote, expected)
		}
	}
}

func TestSyncStatusIsLinked(t *testing.T) {
	if (SyncStatus{State: syncstate.NotLinked}).IsLinked() {
		t.Error("not_linked status reported as linked")
	}
	for _, state := range []enums.SyncState{syncstate.UpToDate, syncstate.Error, ""} {
		if !(SyncStatus{State: state}).IsLinked() {
			t.Errorf("state %q reported as not linked", state)
		}
	}
}

func TestConflictItemChoices(t *testing.T) {
	cases := map[enums.ConflictKind][]enums.ConflictResolution{
		conflictkind.Field:        {conflictresolution.Mine, conflictresolution.Theirs},
		conflictkind.DeleteModify: {conflictresolution.Keep, conflictresolution.Delete},
		conflictkind.BothModified: {conflictresolution.Mine, conflictresolution.Theirs, conflictresolution.Both},
		conflictkind.Unknown:      {conflictresolution.Mine, conflictresolution.Theirs, conflictresolution.Both},
	}
	for kind, expected := range cases {
		if choices := (ConflictItem{Kind: kind}).Choices(); !reflect.DeepEqual(choices, expected) {
			t.Errorf("Choices(%q) = %v, want %v", kind, choices, expected)
		}
	}
}

func TestChosenResolution(t *testing.T) {
	if chosen := (ConflictItem{Resolution: conflictresolution.Unresolved}).ChosenResolution(); chosen != "" {
		t.Errorf("unresolved item chose %q", chosen)
	}
	if chosen := (ConflictItem{Resolution: conflictresolution.Merged}).ChosenResolution(); chosen != conflictresolution.Merged {
		t.Errorf("merged item chose %q", chosen)
	}
	if chosen := (FieldConflict{Resolution: conflictresolution.Unresolved}).ChosenResolution(); chosen != "" {
		t.Errorf("unresolved field chose %q", chosen)
	}
	if chosen := (FieldConflict{Resolution: conflictresolution.Theirs}).ChosenResolution(); chosen != conflictresolution.Theirs {
		t.Errorf("field chose %q", chosen)
	}
}

func TestChangeItemIDs(t *testing.T) {
	if ids := ChangeItemIDs(nil); ids == nil || len(ids) != 0 {
		t.Errorf("ChangeItemIDs(nil) = %#v, want an empty non-nil slice", ids)
	}
	ids := ChangeItemIDs([]ChangeItem{{ItemID: "a"}, {ItemID: "b"}})
	if !reflect.DeepEqual(ids, []string{"a", "b"}) {
		t.Errorf("ChangeItemIDs = %v", ids)
	}
}
