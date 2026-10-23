package entity

import (
	"reflect"
	"testing"
)

func aliasedRequest() Request {
	request := NewRequest("r1", "Login")
	request.Description = "Signs in"
	request.Method = "POST"
	request.URL = "https://api.test/v1/login"
	request.Headers = []KeyValue{NewKeyValue("Accept", "application/json")}
	request.Scripts = Scripts{PreRequest: "default()"}
	request.Examples = []SavedExample{NewSavedExample("e1", "OK")}
	request.Revision = "rev"
	alias := NewRequestAlias("a1", "  mx-test-user ", request, nil)
	alias.Version.URL = "https://api.test/v2/login"
	alias.Version.Scripts = Scripts{PreRequest: "alias()"}
	return request.WithAliasAppended(alias)
}

func TestNewAliasCopiesOnlyTheSendableContent(t *testing.T) {
	alias, _ := aliasedRequest().Alias("a1")
	if alias.Name != "mx-test-user" {
		t.Errorf("Name = %q, want the trimmed name", alias.Name)
	}
	version := alias.Version
	if version.ID != "" || version.Name != "" || version.Description != "" || version.Examples != nil || version.Aliases != nil || version.Revision != "" {
		t.Errorf("an alias keeps only sendable content, got %+v", version)
	}
	if version.Method != "POST" || len(version.Headers) != 1 {
		t.Errorf("the sendable content is copied, got %+v", version)
	}
}

func TestVersionSwapsTheSendableContentAndKeepsTheRequestIdentity(t *testing.T) {
	request := aliasedRequest()
	if got := request.Version(""); !reflect.DeepEqual(got, request) {
		t.Error("the Default version is the request itself")
	}
	version := request.Version("a1")
	if version.URL != "https://api.test/v2/login" || version.Scripts.PreRequest != "alias()" {
		t.Errorf("Version(a1) = %+v, want the alias content", version)
	}
	if version.ID != "r1" || version.Name != "Login" || version.Description != "Signs in" || len(version.Examples) != 1 || version.Revision != "rev" {
		t.Errorf("Version(a1) must keep the request identity, got %+v", version)
	}
}

func TestAliasNamesAreComparedIgnoringCaseAndSpaces(t *testing.T) {
	request := aliasedRequest()
	if !request.HasAliasNamed(" MX-Test-User ", "other") {
		t.Error("a name in another case is taken")
	}
	if request.HasAliasNamed("mx-test-user", "a1") {
		t.Error("an alias does not clash with itself")
	}
	for _, name := range []string{"default", " DEFAULT "} {
		if !(RequestAlias{Name: name}).HasReservedName() {
			t.Errorf("%q is reserved", name)
		}
	}
}

func TestAliasListChangesReturnCopies(t *testing.T) {
	request := aliasedRequest()
	renamed, _ := request.Alias("a1")
	updated := request.WithAlias(renamed.Renamed("qa"))
	if alias, _ := request.Alias("a1"); alias.Name != "mx-test-user" {
		t.Error("WithAlias must not change the original request")
	}
	if alias, _ := updated.Alias("a1"); alias.Name != "qa" {
		t.Error("WithAlias replaces the alias with the same id")
	}
	if removed := updated.WithoutAlias("a1"); len(removed.Aliases) != 0 || len(updated.Aliases) != 1 {
		t.Error("WithoutAlias removes only from the copy")
	}
	if copied := request.CopyAs("r2", "Login Copy"); len(copied.Aliases) != 1 {
		t.Error("a copy of a request carries its aliases")
	}
}

func TestEditedKeepsTheStoredScriptsUnlessTheEditChangedThem(t *testing.T) {
	alias, _ := aliasedRequest().Alias("a1")
	edit := NewRequest("r1", "Login")
	edit.URL = "https://api.test/v3/login"
	if got := alias.Edited(edit).Version; got.URL != "https://api.test/v3/login" || got.Scripts.PreRequest != "alias()" {
		t.Errorf("Edited without edited scripts = %+v, want the stored scripts", got)
	}
	edit.Scripts = Scripts{PreRequest: "edited()"}
	edit.ScriptsEdited = true
	if got := alias.Edited(edit).Version; got.Scripts.PreRequest != "edited()" {
		t.Errorf("Edited with edited scripts = %+v, want them", got)
	}
}

func TestAliasAuthOwnersRoundTrip(t *testing.T) {
	owner := AliasAuthOwner("r1", "a1")
	if requestId, aliasId := SplitAuthOwner(owner); requestId != "r1" || aliasId != "a1" {
		t.Errorf("SplitAuthOwner(%q) = %q, %q", owner, requestId, aliasId)
	}
	if requestId, aliasId := SplitAuthOwner("r1"); requestId != "r1" || aliasId != "" {
		t.Errorf("a plain owner has no alias, got %q, %q", requestId, aliasId)
	}
	effective := EffectiveRequest{AuthOwner: "r1"}
	if got := effective.ForAlias("r1", "a1").AuthOwner; got != owner {
		t.Errorf("an alias's own auth is owned by %q, got %q", owner, got)
	}
	if got := (EffectiveRequest{AuthOwner: "folder"}).ForAlias("r1", "a1").AuthOwner; got != "folder" {
		t.Errorf("an inherited auth keeps its owner, got %q", got)
	}
}

func TestPostmanExportWritesAliasesAsRequestsAfterTheirDefault(t *testing.T) {
	request := aliasedRequest()
	root := NewContainerExportNode(NewCollection("c1", "Billing"), "Billing", []ExportNode{NewRequestExportNode(request, "Billing/Login")})
	children := root.WithAliasesAsRequests(noLabels).Children
	if len(children) != 2 {
		t.Fatalf("children = %d, want the Default and one alias", len(children))
	}
	if children[0].Request.Aliases != nil {
		t.Error("the exported Default has no aliases")
	}
	alias := children[1]
	if alias.Request.Name != "Login (mx-test-user)" || alias.Path != "Billing/Login (mx-test-user)" || alias.Request.ID != "r1-a1" {
		t.Errorf("alias node = %q %q %q", alias.Request.Name, alias.Path, alias.Request.ID)
	}
	if alias.Request.URL != "https://api.test/v2/login" || alias.Request.Examples != nil {
		t.Errorf("an alias request has the alias content and no examples, got %+v", alias.Request)
	}
}

func noLabels(ids []string) []string {
	return ids
}

func TestPostmanExportNotesALinkedAliasesEnvironmentsInItsDescription(t *testing.T) {
	request := NewRequest("r1", "Login")
	request.Description = "Signs in"
	linked := NewRequestAlias("a1", "mx-test-user", request, []string{"hml", "gone"})
	unlinked := NewRequestAlias("a2", "v2", request, nil)
	request = request.WithAliasAppended(linked).WithAliasAppended(unlinked)
	labels := func(ids []string) []string {
		return AliasEnvironmentLabels(ids, []Environment{NewEnvironment("hml", "HML")})
	}

	nodes := NewRequestExportNode(request, "Billing/Login").AliasRequests(labels)
	if got := nodes[0].Request.Description; got != "Signs in\n\nXenon alias for environments: HML, Deleted environment" {
		t.Errorf("linked alias description = %q", got)
	}
	if got := nodes[1].Request.Description; got != "Signs in" {
		t.Errorf("an alias for all environments exports unchanged, got %q", got)
	}

	request.Description = ""
	request.Aliases = []RequestAlias{linked}
	if got := NewRequestExportNode(request, "Billing/Login").AliasRequests(labels)[0].Request.Description; got != "Xenon alias for environments: HML, Deleted environment" {
		t.Errorf("without a description only the note is written, got %q", got)
	}
}
