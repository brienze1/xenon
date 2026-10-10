package entity

import (
	"reflect"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
	"github.com/brienze1/xenon/internal/domain/enums/referencestatus"
	"github.com/brienze1/xenon/internal/domain/enums/variablescope"
	"github.com/brienze1/xenon/internal/domain/enums/vaultmode"
)

func TestVariableRowKeyRules(t *testing.T) {
	cases := map[string]struct{ valid, spaced bool }{
		"base_url":    {true, false},
		" name ":      {true, true},
		"name ":       {true, true},
		"$guid":       {true, false},
		"my-var.v2:x": {true, false},
		"":            {false, false},
		"a{b":         {false, false},
		"a}b":         {false, false},
		"line\nbreak": {false, false},
		"carriage\rx": {false, false},
	}
	for key, want := range cases {
		row := NewVariableRow(key, "v")
		if got := row.HasValidKey(); got != want.valid {
			t.Errorf("HasValidKey(%q) = %v, want %v", key, got, want.valid)
		}
		if got := row.HasSurroundingSpaces(); got != want.spaced {
			t.Errorf("HasSurroundingSpaces(%q) = %v, want %v", key, got, want.spaced)
		}
	}
}

func TestVariableRowActivityAndSecretValues(t *testing.T) {
	row := NewVariableRow("token", "abc")
	if !row.Enabled || row.Secret || row.InitialValue != "abc" || !row.IsActive() {
		t.Fatalf("NewVariableRow = %+v", row)
	}
	if (VariableRow{Key: "token"}).IsActive() {
		t.Error("a disabled row is not active")
	}
	if (VariableRow{Enabled: true}).IsActive() {
		t.Error("a row without a key is not active")
	}
	stray := VariableRow{Key: "api_key", InitialValue: "leaked", Secret: true, Enabled: true}
	if !stray.HasStraySecretValue() || row.HasStraySecretValue() || (VariableRow{Secret: true}).HasStraySecretValue() {
		t.Error("only a secret row holding a value has a stray value")
	}
	if cleaned := stray.WithoutSecretValue(); cleaned.InitialValue != "" || !cleaned.Secret || cleaned.Key != "api_key" {
		t.Errorf("WithoutSecretValue() = %+v", cleaned)
	}
	if kept := row.WithoutSecretValue(); kept != row {
		t.Errorf("a plain row keeps its value, got %+v", kept)
	}
	if logged := stray.LogValue().(map[string]any); logged["key"] != "api_key" || logged["secret"] != true || logged["enabled"] != true || len(logged) != 3 {
		t.Errorf("LogValue() = %v", logged)
	}
}

func TestVariableRowContextVariable(t *testing.T) {
	current := map[string]string{"base_url": "http://localhost", "empty": ""}
	secrets := map[string]string{"api_key": "s3cr3t"}
	cases := map[string]struct {
		row  VariableRow
		want ContextVariable
	}{
		"initial without override":   {NewVariableRow("tenant", "acme"), ContextVariable{Name: "tenant", Value: "acme", Scope: variablescope.Globals}},
		"current overrides initial":  {NewVariableRow("base_url", "https://dev"), ContextVariable{Name: "base_url", Value: "http://localhost", Scope: variablescope.Globals}},
		"empty current is no value":  {NewVariableRow("empty", "initial"), ContextVariable{Name: "empty", Value: "initial", Scope: variablescope.Globals}},
		"secret from the vault":      {VariableRow{Key: "api_key", Secret: true, Enabled: true}, ContextVariable{Name: "api_key", Value: "s3cr3t", Scope: variablescope.Globals, Secret: true}},
		"secret without vault entry": {VariableRow{Key: "password", Secret: true, Enabled: true}, ContextVariable{Name: "password", Scope: variablescope.Globals, Secret: true, Missing: true}},
		"secret ignores file values": {VariableRow{Key: "base_url", InitialValue: "leaked", Secret: true, Enabled: true}, ContextVariable{Name: "base_url", Scope: variablescope.Globals, Secret: true, Missing: true}},
	}
	for name, tc := range cases {
		if got := tc.row.ContextVariable(variablescope.Globals, current, secrets); got != tc.want {
			t.Errorf("%s: got %+v, want %+v", name, got, tc.want)
		}
	}
	variable := ContextVariable{Name: "x", Value: "hidden", Scope: variablescope.Environment, Secret: true, Missing: true}
	if variable.IsResolvable() || !(ContextVariable{}).IsResolvable() {
		t.Error("only missing variables are not resolvable")
	}
	logged := variable.LogValue().(map[string]any)
	if logged["name"] != "x" || logged["scope"] != variablescope.Environment || logged["secret"] != true || logged["missing"] != true || logged["locked"] != false || len(logged) != 5 {
		t.Errorf("LogValue() = %v", logged)
	}
}

func TestVariableLayerAndContextPrecedence(t *testing.T) {
	environment := NewVariableLayer(variablescope.Environment, []VariableRow{
		NewVariableRow("tenant", "dev"),
		{Key: "region", InitialValue: "disabled", Enabled: false},
		NewVariableRow("dup", "first"),
		NewVariableRow("dup", "last"),
		{Key: "dup_disabled", InitialValue: "kept", Enabled: true},
		{Key: "dup_disabled", InitialValue: "ignored", Enabled: false},
		{InitialValue: "no key", Enabled: true},
	}, map[string]string{"tenant": "dev-current"}, nil, false)
	globals := NewVariableLayer(variablescope.Globals, []VariableRow{NewVariableRow("tenant", "global"), NewVariableRow("region", "eu")}, nil, nil, false)
	context := NewVariableContext(environment, globals)
	expectations := map[string]ContextVariable{
		"tenant":       {Name: "tenant", Value: "dev-current", Scope: variablescope.Environment},
		"region":       {Name: "region", Value: "eu", Scope: variablescope.Globals},
		"dup":          {Name: "dup", Value: "last", Scope: variablescope.Environment},
		"dup_disabled": {Name: "dup_disabled", Value: "kept", Scope: variablescope.Environment},
	}
	for name, want := range expectations {
		if got, found := context.Lookup(name); !found || got != want {
			t.Errorf("Lookup(%q) = %+v, %v; want %+v", name, got, found, want)
		}
	}
	for _, name := range []string{"", "missing", "Tenant", " tenant "} {
		if _, found := context.Lookup(name); found {
			t.Errorf("Lookup(%q) must not be found", name)
		}
	}
	if environment.Scope != variablescope.Environment || len(environment.Variables) != 3 {
		t.Errorf("environment layer = %+v", environment)
	}
	if empty := NewVariableContext(); len(empty.Layers) != 0 {
		t.Error("an empty context has no layers")
	}
}

func TestScopeRef(t *testing.T) {
	cases := map[string]ScopeRef{
		"env:01E":        {Scope: variablescope.Environment, ID: "01E"},
		"collection:01C": {Scope: variablescope.Collection, ID: "01C"},
		"globals":        {Scope: variablescope.Globals},
	}
	for raw, want := range cases {
		got := ParseScopeRef(raw)
		if got != want || !got.IsValid() || got.String() != raw {
			t.Errorf("ParseScopeRef(%q) = %+v (%q, valid %v)", raw, got, got.String(), got.IsValid())
		}
	}
	for _, raw := range []string{"", "env:", "collection:", "Globals", "folder:01F", "local"} {
		got := ParseScopeRef(raw)
		if got.IsValid() || got.Scope != variablescope.Unknown || got.String() != raw {
			t.Errorf("ParseScopeRef(%q) = %+v must be invalid and keep the raw text", raw, got)
		}
	}
	if NewEnvironmentScopeRef("e").String() != "env:e" || NewCollectionScopeRef("c").String() != "collection:c" || NewGlobalsScopeRef().String() != "globals" {
		t.Error("constructors and String disagree")
	}
}

func TestLocalVariables(t *testing.T) {
	local := NewLocalVariables()
	if local.Current == nil || local.HasActiveEnvironment() || local.CurrentValues(NewGlobalsScopeRef()) != nil {
		t.Fatalf("NewLocalVariables() = %+v", local)
	}
	local.Current["env:e1"] = map[string]string{"base_url": "http://localhost"}
	active := local.WithActiveEnvironment("e1")
	if !active.HasActiveEnvironment() || active.ActiveEnvironmentID != "e1" || local.HasActiveEnvironment() {
		t.Errorf("WithActiveEnvironment returns a copy, got %+v and %+v", active, local)
	}
	if got := active.CurrentValues(NewEnvironmentScopeRef("e1")); got["base_url"] != "http://localhost" {
		t.Errorf("CurrentValues = %v", got)
	}
	if logged := active.LogValue().(map[string]any); logged["activeEnvironmentId"] != "e1" || logged["scopes"] != 1 || len(logged) != 2 {
		t.Errorf("LogValue() = %v", logged)
	}
}

func TestNewEnvironmentDefaultsProductionNames(t *testing.T) {
	for name, production := range map[string]bool{"prod": true, " Production ": true, "pre-PROD": true, "dev": false, "pro": false} {
		environment := NewEnvironment("01E", name)
		if environment.Name != strings.TrimSpace(name) || environment.ID != "01E" || environment.Variables == nil {
			t.Errorf("NewEnvironment(%q) = %+v", name, environment)
		}
		wantColor := envcolor.None
		if production {
			wantColor = envcolor.Red
		}
		if environment.Color != wantColor || environment.ConfirmBeforeSend != production || environment.IsProductionLike() != production {
			t.Errorf("NewEnvironment(%q) color %q confirm %v", name, environment.Color, environment.ConfirmBeforeSend)
		}
	}
}

func TestEnvironmentPredicates(t *testing.T) {
	environment := NewEnvironment("01E", "Dev")
	if !environment.HasID("01E") || environment.HasID("01X") || environment.IsMalformed() {
		t.Error("a readable environment matches its own id")
	}
	if !environment.HasName(" dev ") || environment.HasName("hml") {
		t.Error("names compare case-insensitively without surrounding spaces")
	}
	if !environment.HasValidName() || NewEnvironment("01E", " ").HasValidName() || NewEnvironment("01E", strings.Repeat("a", 201)).HasValidName() {
		t.Error("names need 1 to 200 characters")
	}
	if environment.ScopeRef() != NewEnvironmentScopeRef("01E") {
		t.Errorf("ScopeRef() = %+v", environment.ScopeRef())
	}
	malformed := NewMalformedEnvironment("broken.yaml", "cannot be read")
	if !malformed.IsMalformed() || malformed.HasID(malformed.ID) || malformed.ID != "error:broken.yaml" || malformed.Name != "broken.yaml" || malformed.Color != envcolor.None || malformed.FileName != "broken.yaml" || malformed.Variables == nil {
		t.Errorf("NewMalformedEnvironment = %+v", malformed)
	}
	logged := environment.LogValue().(map[string]any)
	if logged["id"] != "01E" || logged["name"] != "Dev" || logged["color"] != envcolor.None || logged["variables"] != 0 || logged["error"] != "" || len(logged) != 5 {
		t.Errorf("LogValue() = %v", logged)
	}
}

func TestNewEnvironmentListSortsAndChecksTheActiveId(t *testing.T) {
	prod := NewEnvironment("p", "Prod")
	dev := NewEnvironment("d", "dev")
	twinA := Environment{ID: "a", Name: "same", FileName: "same.yaml"}
	twinB := Environment{ID: "b", Name: "same", FileName: "same.yaml"}
	sameNameLater := Environment{ID: "c", Name: "same", FileName: "z.yaml"}
	broken := NewMalformedEnvironment("broken.yaml", "bad")
	input := []Environment{prod, sameNameLater, twinA, twinB, broken, dev}
	list := NewEnvironmentList(input, "d")
	ids := make([]string, 0, len(list.Environments))
	for _, environment := range list.Environments {
		ids = append(ids, environment.ID)
	}
	if want := []string{"error:broken.yaml", "d", "p", "a", "b", "c"}; !reflect.DeepEqual(ids, want) {
		t.Errorf("order = %v, want %v", ids, want)
	}
	if input[0].ID != "p" {
		t.Error("the input slice is not reordered")
	}
	if list.ActiveID != "d" || !list.Contains("p") || list.Contains("x") || list.Contains(broken.ID) {
		t.Errorf("list = %+v", list)
	}
	for _, active := range []string{"", "missing", broken.ID} {
		if got := NewEnvironmentList(input, active).ActiveID; got != "" {
			t.Errorf("active %q must be dropped, got %q", active, got)
		}
	}
	if logged := list.LogValue().(map[string]any); logged["activeId"] != "d" || logged["environments"] != 6 || len(logged) != 2 {
		t.Errorf("LogValue() = %v", logged)
	}
}

func TestReferenceAndResolvedReference(t *testing.T) {
	reference := Reference{Name: "base_url", Field: "url", Start: 0, End: 12}
	if reference.Placeholder() != "{{base_url}}" || reference.IsDynamic() || !(Reference{Name: "$guid"}).IsDynamic() {
		t.Errorf("reference = %+v", reference)
	}
	cases := []struct {
		reference ResolvedReference
		resolved  bool
		display   string
	}{
		{ResolvedReference{Value: "v", Status: referencestatus.Resolved}, true, "v"},
		{ResolvedReference{Value: "s3cr3t", Status: referencestatus.Resolved, Secret: true}, true, MaskedValue},
		{ResolvedReference{Status: referencestatus.Dynamic, Value: "generated"}, true, ""},
		{ResolvedReference{Status: referencestatus.Unresolved, Value: "{{x}}"}, false, "{{x}}"},
		{ResolvedReference{Status: referencestatus.MissingSecret, Secret: true}, false, MaskedValue},
		{ResolvedReference{Status: referencestatus.LockedSecret, Secret: true}, false, MaskedValue},
	}
	for _, tc := range cases {
		if tc.reference.IsResolved() != tc.resolved || tc.reference.DisplayValue() != tc.display {
			t.Errorf("%+v: resolved %v display %q", tc.reference, tc.reference.IsResolved(), tc.reference.DisplayValue())
		}
	}
}

func TestSendPreflight(t *testing.T) {
	confirm := NewEnvironment("p", "prod")
	cases := []struct {
		preflight                      SendPreflight
		confirmation, secrets, isClear bool
	}{
		{SendPreflight{}, false, false, true},
		{SendPreflight{ConfirmEnvironment: &confirm}, true, false, false},
		{SendPreflight{MissingSecrets: []string{"api_key"}}, false, true, false},
		{SendPreflight{LockedSecrets: []string{"password"}}, false, true, false},
		{SendPreflight{ConfirmEnvironment: &confirm, MissingSecrets: []string{"a"}, LockedSecrets: []string{"b"}}, true, true, false},
	}
	for _, tc := range cases {
		p := tc.preflight
		if p.NeedsConfirmation() != tc.confirmation || p.HasUnavailableSecrets() != tc.secrets || p.IsClear() != tc.isClear {
			t.Errorf("%+v: confirm %v secrets %v clear %v", p, p.NeedsConfirmation(), p.HasUnavailableSecrets(), p.IsClear())
		}
	}
}

func TestVaultStatus(t *testing.T) {
	status := VaultStatus{Mode: vaultmode.Passphrase, LockedLocalIDs: []string{"a"}, BrokenLocalIDs: []string{"b"}}
	if !status.IsLocked("a") || status.IsLocked("b") || !status.IsBroken("b") || status.IsBroken("a") {
		t.Errorf("status = %+v", status)
	}
}

func TestCollectionAndWorkspaceVariables(t *testing.T) {
	collection := NewCollection("c", "Billing")
	if collection.Variables == nil || NewWorkspace("w", "Acme").Variables == nil {
		t.Fatal("new collections and workspaces start with no variables")
	}
	collection.Variables = []VariableRow{NewVariableRow("page_size", "50")}
	if got := collection.OwnVariables(); !reflect.DeepEqual(got, collection.Variables) {
		t.Errorf("collection OwnVariables() = %v", got)
	}
	folder := NewFolder("f", "Invoices")
	folder.Variables = []VariableRow{NewVariableRow("ignored", "x")}
	if got := folder.OwnVariables(); got == nil || len(got) != 0 {
		t.Errorf("folders have no variables, got %v", got)
	}
	if collection.ScopeRef() != NewCollectionScopeRef("c") {
		t.Errorf("ScopeRef() = %+v", collection.ScopeRef())
	}
	settings := NewCollection("c", "Billing")
	settings.Description = "new"
	if kept := collection.WithSettings(settings); !reflect.DeepEqual(kept.Variables, collection.Variables) {
		t.Error("settings never replace variables")
	}
}

func TestMapTextWalksEveryResolvableField(t *testing.T) {
	mark := func(text string) string {
		if text == "" {
			return ""
		}
		return "<" + text + ">"
	}
	request := NewRequest("r", "Request")
	request.Name = "name"
	request.Description = "description"
	request.Method = "POST"
	request.URL = "url"
	request.Params = []KeyValue{{Key: "pk", Value: "pv", Description: "pd", Enabled: true}}
	request.PathVariables = []PathVariable{{Key: "id", Value: "pathValue", Description: "d"}}
	request.Headers = []KeyValue{NewKeyValue("hk", "hv")}
	request.Auth = Auth{Type: authtype.Bearer, Bearer: AuthBearer{Token: "token"}, Basic: AuthBasic{Username: "user", Password: "pass"}}
	request.Body = Body{
		Mode:       bodymode.Raw,
		Raw:        BodyRaw{Content: "content"},
		GraphQl:    BodyGraphQl{Query: "query", Variables: "variables"},
		Binary:     FileRef{Src: "binary"},
		FormData:   []FormDataRow{{Key: "fk", Value: "fv", Src: FileRef{Src: "file"}, Description: "fd", Enabled: true}},
		UrlEncoded: []KeyValue{NewKeyValue("uk", "uv")},
	}
	mapped := request.MapText(mark)
	want := request
	want.URL = "<url>"
	want.Params = []KeyValue{{Key: "<pk>", Value: "<pv>", Description: "pd", Enabled: true}}
	want.PathVariables = []PathVariable{{Key: "id", Value: "<pathValue>", Description: "d"}}
	want.Headers = []KeyValue{NewKeyValue("<hk>", "<hv>")}
	want.Auth = Auth{Type: authtype.Bearer, Bearer: AuthBearer{Token: "<token>"}, Basic: AuthBasic{Username: "<user>", Password: "<pass>"}}
	want.Body = Body{
		Mode:       bodymode.Raw,
		Raw:        BodyRaw{Content: "<content>"},
		GraphQl:    BodyGraphQl{Query: "<query>", Variables: "<variables>"},
		Binary:     FileRef{Src: "<binary>"},
		FormData:   []FormDataRow{{Key: "<fk>", Value: "<fv>", Src: FileRef{Src: "<file>"}, Description: "fd", Enabled: true}},
		UrlEncoded: []KeyValue{NewKeyValue("<uk>", "<uv>")},
	}
	if !reflect.DeepEqual(mapped, want) {
		t.Fatalf("MapText() =\n%+v\nwant\n%+v", mapped, want)
	}
	if request.URL != "url" || request.Headers[0].Key != "hk" || request.Body.FormData[0].Key != "fk" || request.PathVariables[0].Value != "pathValue" {
		t.Error("MapText works on a copy")
	}
	collection := NewCollection("c", "C")
	collection.Auth = Auth{Type: authtype.Basic, Basic: AuthBasic{Username: "u", Password: "p"}}
	collection.Headers = []KeyValue{NewKeyValue("ck", "cv")}
	collection.Variables = []VariableRow{NewVariableRow("k", "v")}
	mappedCollection := collection.MapText(mark)
	if mappedCollection.Auth.Basic != (AuthBasic{Username: "<u>", Password: "<p>"}) || mappedCollection.Headers[0] != NewKeyValue("<ck>", "<cv>") || mappedCollection.Variables[0] != NewVariableRow("k", "v") || mappedCollection.Name != "C" {
		t.Errorf("collection MapText() = %+v", mappedCollection)
	}
	if collection.Headers[0].Key != "ck" {
		t.Error("collection MapText works on a copy")
	}
	empty := NewRequest("", "").MapText(mark)
	if empty.Params == nil || empty.PathVariables == nil || empty.Headers == nil || empty.Body.FormData == nil || empty.Body.UrlEncoded == nil {
		t.Errorf("mapping keeps empty lists non-nil: %+v", empty)
	}
}
