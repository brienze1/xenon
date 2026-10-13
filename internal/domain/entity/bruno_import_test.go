package entity

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/importitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

func TestScriptTranslationAppendKeepsOrderWithoutSharing(t *testing.T) {
	first := ScriptTranslation{Lines: []string{"a"}, Untranslated: []UntranslatedLine{{Line: 1, Source: "a"}}}
	joined := first.Append(ScriptTranslation{Lines: []string{"b", "c"}, Untranslated: []UntranslatedLine{{Line: 3, Source: "c"}}})
	if joined.Text() != "a\nb\nc" || len(joined.Untranslated) != 2 || joined.Untranslated[1].Line != 3 {
		t.Fatalf("joined = %+v", joined)
	}
	joined.Lines[0] = "changed"
	if first.Lines[0] != "a" || (ScriptTranslation{}).Text() != "" {
		t.Error("append must copy")
	}
}

func TestBrunoAssertionSourceAndTitle(t *testing.T) {
	cases := map[BrunoAssertion][2]string{
		{Expression: "res.status", Operator: "eq", Value: "200"}: {"res.status: eq 200", "res.status eq 200"},
		{Expression: "res.body", Operator: "isJson"}:             {"res.body: isJson", "res.body isJson"},
		{Expression: "res.body", Operator: "", Value: ""}:        {"res.body:", "res.body"},
	}
	for assertion, want := range cases {
		if assertion.Source() != want[0] || assertion.Title() != want[1] {
			t.Errorf("%+v: %q %q", assertion, assertion.Source(), assertion.Title())
		}
	}
}

func TestBrunoScriptsTranslatedReportsEveryUntranslatedLine(t *testing.T) {
	scripts := BrunoScripts{ItemID: "id", Path: "Shop/A", File: "A.bru"}
	pre := ScriptTranslation{Lines: []string{"p"}, Untranslated: []UntranslatedLine{{Line: 4, Source: "bru.x();"}}}
	post := ScriptTranslation{Lines: []string{"q", "r"}, Untranslated: []UntranslatedLine{{Line: 9, Source: "res(1)"}}}
	translated, warnings := scripts.Translated(pre, post)
	if translated.PreRequest != "p" || translated.PostResponse != "q\nr" {
		t.Errorf("scripts = %+v", translated)
	}
	if len(warnings) != 2 || warnings[0].Code != warningcode.UntranslatedScript || warnings[0].Location != "A.bru:4" || warnings[1].Location != "A.bru:9" {
		t.Fatalf("warnings = %+v", warnings)
	}
	if warnings[0].Message != "This line has no pm.* equivalent and was commented out: bru.x();" || warnings[0].ItemID != "id" || warnings[0].Path != "Shop/A" {
		t.Errorf("warning = %+v", warnings[0])
	}
}

func TestBrunoScriptsIsEmptyOnlyWithoutAnyScriptPart(t *testing.T) {
	cases := map[string]struct {
		scripts BrunoScripts
		empty   bool
	}{
		"nothing":    {BrunoScripts{PreRequest: BrunoScript{Text: " \n"}}, true},
		"pre":        {BrunoScripts{PreRequest: BrunoScript{Text: "a"}}, false},
		"post":       {BrunoScripts{PostResponse: BrunoScript{Text: "a"}}, false},
		"tests":      {BrunoScripts{Tests: BrunoScript{Text: "a"}}, false},
		"pre vars":   {BrunoScripts{PreRequestVars: []BrunoVariable{{}}}, false},
		"post vars":  {BrunoScripts{PostResponseVars: []BrunoVariable{{}}}, false},
		"assertions": {BrunoScripts{Assertions: []BrunoAssertion{{}}}, false},
	}
	for name, tc := range cases {
		if tc.scripts.IsEmpty() != tc.empty {
			t.Errorf("%s: IsEmpty = %v", name, !tc.empty)
		}
	}
}

func TestImportNodeWithScriptsSetsScriptsByItemId(t *testing.T) {
	folder := NewFolder("f", "F")
	request := NewRequest("r", "R")
	other := NewRequest("o", "O")
	root := NewContainerImportNode(NewCollection("c", "C"), []ImportNode{NewContainerImportNode(folder, []ImportNode{NewRequestImportNode(request), NewRequestImportNode(other)})})
	scripts := map[string]Scripts{"c": {PreRequest: "c"}, "f": {PostResponse: "f"}, "r": {PreRequest: "r"}}
	updated := root.WithScripts(scripts)
	if updated.Container.Scripts.PreRequest != "c" || updated.Children[0].Container.Scripts.PostResponse != "f" {
		t.Errorf("containers = %+v", updated)
	}
	if updated.Children[0].Children[0].Request.Scripts.PreRequest != "r" || !updated.Children[0].Children[1].Request.Scripts.IsEmpty() {
		t.Errorf("requests = %+v", updated.Children[0].Children)
	}
	if root.Children[0].Children[0].Request.Scripts.PreRequest != "" || root.Container.Scripts.PreRequest != "" {
		t.Error("the original tree must not change")
	}
}

func TestImportNodePlacedIntoKeepsARequestAsItIs(t *testing.T) {
	parent := NewFolder("p", "P")
	parent.Headers = []KeyValue{NewKeyValue("a", "b")}
	placement := NewRequestImportNode(NewRequest("r", "R")).PlacedInto(parent)
	if placement.Merge || placement.Wrapped || placement.Parent.ID != "p" || len(placement.Nodes) != 1 || !placement.Nodes[0].IsRequest() {
		t.Errorf("request placement = %+v", placement)
	}
	wrapped := NewContainerImportNode(NewCollection("c", "C"), []ImportNode{}).PlacedInto(parent)
	if !wrapped.Wrapped || wrapped.Merge || wrapped.Nodes[0].Kind() != itemkind.Folder {
		t.Errorf("wrapped placement = %+v", wrapped)
	}
	merged := NewContainerImportNode(NewCollection("c", "C"), []ImportNode{}).PlacedInto(NewFolder("e", "E"))
	if merged.Wrapped || !merged.Merge {
		t.Errorf("merged placement = %+v", merged)
	}
}

func TestImportItemNamedRenamesTreesAndEnvironments(t *testing.T) {
	item := NewImportItem("0", importitemkind.Request, NewRequestImportNode(NewRequest("r", "GET /x")), []ImportWarning{})
	if !item.IsRequest() || item.Named("  ").Name != "GET /x" {
		t.Fatalf("item = %+v", item)
	}
	renamed := item.Named(" Invite ")
	if renamed.Name != "Invite" || renamed.Root.Name() != "Invite" || item.Root.Name() != "GET /x" {
		t.Errorf("renamed = %+v", renamed)
	}
	variables := NewVariablesImportItem("1", importitemkind.Environment, NewImportVariables(NewEnvironment("e", "dev"), map[string]string{}))
	renamedVariables := variables.Named("qa")
	if variables.IsRequest() || renamedVariables.Name != "qa" || renamedVariables.Variables.Environment.Name != "qa" || variables.Variables.Environment.Name != "dev" {
		t.Errorf("variables = %+v", renamedVariables)
	}
}

func TestImportItemCompletedRequestHasNoPlacementWarning(t *testing.T) {
	parent := NewFolder("p", "P")
	node := NewRequestImportNode(NewRequest("r", "R")).WithFileName("r.yaml")
	item := NewImportItem("0", importitemkind.Request, node, []ImportWarning{})
	write := ImportWrite{Parent: &parent, Nodes: []ImportNode{node}}
	outcome := item.Completed(write, ImportLocation{LocalID: "ws", ParentPath: "Main/P"})
	if len(outcome.Warnings) != 0 || outcome.Item.Name != "R" || outcome.Item.ID != "r" {
		t.Errorf("outcome = %+v", outcome)
	}
	write.Wrapped = true
	if wrapped := item.Completed(write, ImportLocation{LocalID: "ws", ParentPath: "Main/P"}); len(wrapped.Warnings) != 1 || wrapped.Warnings[0].Code != warningcode.WrappedInFolder {
		t.Errorf("wrapped = %+v", wrapped.Warnings)
	}
}

func TestParsedCurlLogValueHidesHeaders(t *testing.T) {
	request := NewRequest("", "")
	request.URL = "https://x"
	request.Headers = []KeyValue{NewKeyValue("Authorization", "secret")}
	logged := ParsedCurl{Request: request, Notices: []string{"n"}}.LogValue()
	if !reflect.DeepEqual(logged, map[string]any{"request": request.LogValue(), "notices": []string{"n"}}) {
		t.Errorf("log = %+v", logged)
	}
}

func TestImportProbeIsPastedCurlAcceptsCurlAndCurlExe(t *testing.T) {
	cases := map[string]bool{
		"curl https://x":      true,
		"  CURL.EXE 'x'":      true,
		"curlx https://x":     false,
		"curl":                false,
		"wget https://x/curl": false,
	}
	for text, want := range cases {
		if got := (ImportProbe{}).Pasted(text).IsPastedCurl(); got != want {
			t.Errorf("%q: got %v", text, got)
		}
	}
	if (ImportProbe{Text: "curl https://x"}).IsPastedCurl() {
		t.Error("a file probe is never a pasted cURL command")
	}
}
