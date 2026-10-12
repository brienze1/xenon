package usecase

import (
	"errors"
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/conflictresolution"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestMergeResolvedStructuralFileSettlesTheChosenSide(t *testing.T) {
	base := doc("name", "Billing", "description", "d", "variables", docRows(docRow("v1", "1"), docRow("v2", "1")), "custom", "x")
	mine := doc("name", "Mine", "description", "mine", "variables", docRows(docRow("v1", "2"), docRow("v2", "2")), "custom", "y")
	theirs := doc("name", "Theirs", "description", "theirs", "variables", docRows(docRow("v1", "3")), "custom", "z")
	resolutions := map[string]enums.ConflictResolution{"name": conflictresolution.Theirs, "description": conflictresolution.Mine, "variables[v1].value": conflictresolution.Theirs, "variables[v2]": conflictresolution.Theirs}
	merged := MergeResolvedStructuralFile(itemkind.Collection, base, mine, theirs, resolutions)
	if merged.HasConflicts() {
		t.Fatalf("every conflict was resolved: %+v", merged.FieldConflicts)
	}
	want := doc("name", "Theirs", "description", "mine", "variables", docRows(docRow("v1", "3")), "custom", "z")
	if !merged.Result.Equals(want) {
		t.Fatalf("result %+v", merged.Result)
	}
	partial := MergeResolvedStructuralFile(itemkind.Collection, base, mine, theirs, map[string]enums.ConflictResolution{"name": conflictresolution.Mine})
	paths := make([]string, 0)
	for _, conflict := range partial.FieldConflicts {
		paths = append(paths, conflict.Path)
	}
	if !reflect.DeepEqual(paths, []string{"description", "variables[v1].value", "variables[v2]"}) || partial.Result.Field("name").Value != "Mine" {
		t.Fatalf("unresolved paths %v", paths)
	}
	if MergeStructuralFile(itemkind.Collection, base, mine, theirs).Result.Field("name").Value != "Mine" {
		t.Fatal("no resolutions keeps the mine placeholder")
	}
	environment := MergeResolvedStructuralFile(itemkind.Environment, base, mine, theirs, resolutions)
	if !environment.ItemConflict {
		t.Fatal("environments stay item conflicts")
	}
}

func TestSyncConflictRejectionsCarryTheirCodes(t *testing.T) {
	for code, err := range map[string]error{
		"GIT-04409": RejectUnresolvedConflicts(),
		"GIT-19400": RejectConflictResolution("Users", "both", ""),
		"GIT-35404": RejectMissingConflict("x"),
		"GIT-28409": RejectStaleSyncSession(),
	} {
		var custom errs.BaseError
		if !errors.As(err, &custom) || custom.Code() != code {
			t.Errorf("%s: got %v", code, err)
		}
	}
	var named, unnamed errs.BaseError
	if !errors.As(RejectConflictResolution("API", "both", "name"), &named) || !errors.As(RejectConflictResolution("API", "both", ""), &unnamed) {
		t.Fatal("custom errors")
	}
	if named.Description() != `"both" does not resolve the conflict of API (field "name")` || unnamed.Description() != `"both" does not resolve the conflict of API` {
		t.Fatalf("descriptions %q / %q", named.Description(), unnamed.Description())
	}
}
