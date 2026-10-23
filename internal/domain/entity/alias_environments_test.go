package entity

import (
	"reflect"
	"testing"
)

func TestNormalizedEnvironmentIDsTrimsDropsBlanksAndDuplicatesInFirstOrder(t *testing.T) {
	got := NormalizedEnvironmentIDs([]string{" dev ", "", "hml", "dev", "  "})
	if want := []string{"dev", "hml"}; !reflect.DeepEqual(got, want) {
		t.Errorf("NormalizedEnvironmentIDs = %v, want %v", got, want)
	}
	if got := NormalizedEnvironmentIDs([]string{" ", ""}); got != nil {
		t.Errorf("only blanks normalize to nil, got %v", got)
	}
	if got := NormalizedEnvironmentIDs(nil); got != nil {
		t.Errorf("nil normalizes to nil, got %v", got)
	}
}

func TestAliasEnvironmentLabelsNamesEachIdAndFlagsDeletedOnes(t *testing.T) {
	environments := []Environment{NewEnvironment("hml", "HML"), NewEnvironment("dev", "DEV")}
	got := AliasEnvironmentLabels([]string{"dev", "gone", "hml"}, environments)
	if want := []string{"DEV", "Deleted environment", "HML"}; !reflect.DeepEqual(got, want) {
		t.Errorf("AliasEnvironmentLabels = %v, want %v", got, want)
	}
}

func TestAnUnlinkedAliasIsAvailableEverywhere(t *testing.T) {
	alias := NewRequestAlias("a1", "v2", NewRequest("r1", "Login"), nil)
	hml := NewEnvironment("hml", "HML")
	if alias.IsLinked() {
		t.Error("an alias without environments is not linked")
	}
	if !alias.IsAvailableIn(nil) || !alias.IsAvailableIn(&hml) {
		t.Error("an alias for all environments is available with or without an active environment")
	}
}

func TestALinkedAliasIsAvailableOnlyWhileOneOfItsEnvironmentsIsActive(t *testing.T) {
	alias := NewRequestAlias("a1", "qa", NewRequest("r1", "Login"), []string{"hml", " dev ", "hml"})
	hml, dev, prod := NewEnvironment("hml", "HML"), NewEnvironment("dev", "DEV"), NewEnvironment("prod", "Prod")
	if !reflect.DeepEqual(alias.EnvironmentIDs, []string{"hml", "dev"}) {
		t.Errorf("EnvironmentIDs = %v, want the normalized ids", alias.EnvironmentIDs)
	}
	if !alias.IsAvailableIn(&hml) || !alias.IsAvailableIn(&dev) {
		t.Error("a linked alias is available in each of its environments")
	}
	if alias.IsAvailableIn(&prod) {
		t.Error("a linked alias is unavailable in another environment")
	}
	if alias.IsAvailableIn(nil) {
		t.Error("a linked alias is unavailable with no active environment")
	}
}

func TestWithEnvironmentsNormalizesAndEditsKeepTheLink(t *testing.T) {
	alias := NewRequestAlias("a1", "qa", NewRequest("r1", "Login"), nil).WithEnvironments([]string{"prod", "prod"})
	if !reflect.DeepEqual(alias.EnvironmentIDs, []string{"prod"}) {
		t.Errorf("WithEnvironments = %v, want [prod]", alias.EnvironmentIDs)
	}
	edit := NewRequest("", "")
	edit.URL = "https://api.test/v9"
	edited := alias.Edited(edit).Renamed("renamed")
	if !reflect.DeepEqual(edited.EnvironmentIDs, []string{"prod"}) {
		t.Errorf("Edited/Renamed must keep the link, got %v", edited.EnvironmentIDs)
	}
	if cleared := alias.WithEnvironments([]string{}); cleared.IsLinked() {
		t.Error("an empty set means all environments")
	}
}
