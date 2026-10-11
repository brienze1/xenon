package entity

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/exportsecrets"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
	"github.com/brienze1/xenon/internal/domain/enums/importorigin"
)

func TestExportRequestCombinations(t *testing.T) {
	cases := []struct {
		kind    enums.ExportItemKind
		format  enums.ExportFormat
		shape   enums.ExportShape
		secrets bool
		valid   bool
	}{
		{exportitemkind.Workspace, exportformat.Postman, exportshape.Zip, false, true},
		{exportitemkind.Workspace, exportformat.XenonYaml, exportshape.Zip, true, true},
		{exportitemkind.Workspace, exportformat.Postman, exportshape.File, false, false},
		{exportitemkind.Collection, exportformat.Postman, exportshape.File, false, true},
		{exportitemkind.Collection, exportformat.Postman, exportshape.File, true, false},
		{exportitemkind.Collection, exportformat.Postman, exportshape.Zip, false, false},
		{exportitemkind.Collection, exportformat.XenonYaml, exportshape.Bundle, false, true},
		{exportitemkind.Folder, exportformat.XenonYaml, exportshape.Zip, false, true},
		{exportitemkind.Folder, exportformat.XenonYaml, exportshape.File, false, false},
		{exportitemkind.Request, exportformat.XenonYaml, exportshape.File, false, true},
		{exportitemkind.Request, exportformat.Postman, exportshape.Bundle, false, false},
		{exportitemkind.Environment, exportformat.Postman, exportshape.File, false, true},
		{exportitemkind.Globals, exportformat.Postman, exportshape.File, false, true},
		{exportitemkind.Globals, exportformat.XenonYaml, exportshape.File, false, false},
		{exportitemkind.Unknown, exportformat.Postman, exportshape.File, false, false},
		{exportitemkind.Request, exportformat.Unknown, exportshape.File, false, false},
		{exportitemkind.Workspace, exportformat.Unknown, exportshape.Zip, false, false},
	}
	for _, tc := range cases {
		request := ExportRequest{ItemKind: tc.kind, Format: tc.format, Shape: tc.shape, IncludeSecrets: tc.secrets}
		if got := request.HasValidCombination(); got != tc.valid {
			t.Errorf("%s/%s/%s secrets=%v: valid = %v, want %v", tc.kind, tc.format, tc.shape, tc.secrets, got, tc.valid)
		}
	}
}

func TestExportRequestPassphrase(t *testing.T) {
	cases := []struct {
		request ExportRequest
		valid   bool
	}{
		{ExportRequest{}, true},
		{ExportRequest{IncludeSecrets: true, Passphrase: "12345678", PassphraseConfirmation: "12345678"}, true},
		{ExportRequest{IncludeSecrets: true, Passphrase: "ñññññññ", PassphraseConfirmation: "ñññññññ"}, false},
		{ExportRequest{IncludeSecrets: true, Passphrase: "1234567", PassphraseConfirmation: "1234567"}, false},
		{ExportRequest{IncludeSecrets: true, Passphrase: "12345678", PassphraseConfirmation: "12345679"}, false},
	}
	for _, tc := range cases {
		if got := tc.request.HasValidPassphrase(); got != tc.valid {
			t.Errorf("%+v: valid = %v", tc.request, got)
		}
	}
	request := ExportRequest{Format: exportformat.Postman, ItemKind: exportitemkind.Workspace, Passphrase: "secret-pass"}
	if !request.IsPostman() || !request.IsWorkspace() {
		t.Error("expected a Postman workspace export")
	}
	if logged, _ := request.LogValue().(map[string]any); logged["passphrase"] != nil || len(logged) != 8 {
		t.Errorf("log value = %v", logged)
	}
	if (ExportRequest{Format: exportformat.XenonYaml, ItemKind: exportitemkind.Folder}).IsPostman() {
		t.Error("a Xenon YAML export is not a Postman export")
	}
}

func TestExportFormatOptions(t *testing.T) {
	cases := []struct {
		kind  enums.ExportItemKind
		names []string
	}{
		{exportitemkind.Workspace, []string{"acme.zip", "acme-postman.zip"}},
		{exportitemkind.Collection, []string{"acme.zip", "acme.postman_collection.json"}},
		{exportitemkind.Folder, []string{"acme.zip", "acme.postman_collection.json"}},
		{exportitemkind.Request, []string{"acme.yaml", "acme.postman_collection.json"}},
		{exportitemkind.Environment, []string{"acme.yaml", "acme.postman_environment.json"}},
		{exportitemkind.Globals, []string{"acme.postman_globals.json"}},
		{exportitemkind.Unknown, []string{}},
	}
	for _, tc := range cases {
		names := make([]string, 0)
		for _, option := range ExportFormatOptions(tc.kind, "acme") {
			names = append(names, option.SuggestedFileName)
			if option.IncludeSecretsAllowed != (option.Format == exportformat.XenonYaml) {
				t.Errorf("%s: %+v", tc.kind, option)
			}
		}
		if !reflect.DeepEqual(names, tc.names) {
			t.Errorf("%s: names = %v, want %v", tc.kind, names, tc.names)
		}
	}
	if name := ExportFileName("acme", exportitemkind.Collection, exportformat.XenonYaml, exportshape.Bundle); name != "acme.yaml" {
		t.Errorf("bundle name = %q", name)
	}
}

func TestExportNodeCountsAndNodes(t *testing.T) {
	collection := NewCollection("c", "Billing")
	collection.Variables = []VariableRow{NewVariableRow("a", "1"), NewVariableRow("b", "2")}
	folder := NewFolder("f", "Invoices")
	folder.Variables = []VariableRow{NewVariableRow("ignored", "x")}
	request := NewRequest("r", "Create")
	request.Examples = []SavedExample{NewSavedExample("e", "Ok")}
	root := NewContainerExportNode(collection, "Billing", []ExportNode{
		NewContainerExportNode(folder, "Billing/Invoices", []ExportNode{NewRequestExportNode(request, "Billing/Invoices/Create")}),
		NewRequestExportNode(NewRequest("s", "List"), "Billing/List"),
	})
	want := ExportCounts{Collections: 1, Folders: 1, Requests: 2, Examples: 1, Variables: 2}
	if got := root.Counts(); got != want {
		t.Errorf("counts = %+v, want %+v", got, want)
	}
	names := make([]string, 0)
	for _, node := range root.Nodes() {
		names = append(names, node.Name())
	}
	if !reflect.DeepEqual(names, []string{"Billing", "Invoices", "Create", "List"}) {
		t.Errorf("nodes = %v", names)
	}
	if total := want.Add(ExportCounts{Environments: 2, Variables: 1}); total.Environments != 2 || total.Variables != 3 || total.Requests != 2 {
		t.Errorf("add = %+v", total)
	}
}

func TestExportSummaryAccumulates(t *testing.T) {
	summary := NewExportSummary("/x.json", exportsecrets.NamesOnly).
		With(ExportCounts{Requests: 1}, []ExportDrop{NewExportDrop("A", "color", "r")}, nil).
		With(ExportCounts{Requests: 2}, nil, []ExportWarning{NewExportWarning("literal_credential", "A/B", "m")})
	if summary.Counts.Requests != 3 || len(summary.Dropped) != 1 || len(summary.Warnings) != 1 || summary.Path != "/x.json" || summary.Secrets != exportsecrets.NamesOnly {
		t.Errorf("summary = %+v", summary)
	}
}

func TestExportAuthAndVariableRules(t *testing.T) {
	if !(TextField{Name: "a", Text: "x"}).IsLiteral() || (TextField{Name: "a", Text: "{{x}}"}).IsLiteral() || (TextField{Name: "a"}).IsLiteral() {
		t.Error("IsLiteral")
	}
	if !(TextField{Text: "{{vault:key}}"}).HasVaultReference() || (TextField{Text: "{{key}}"}).HasVaultReference() {
		t.Error("HasVaultReference")
	}
	bruno := NewUnsupportedAuth(AuthUnsupported{Origin: importorigin.Bruno, Type: "wsse"})
	postman := NewUnsupportedAuth(AuthUnsupported{Origin: importorigin.Postman, Type: "hawk"})
	if !bruno.IsBrunoOnly() || postman.IsBrunoOnly() {
		t.Error("IsBrunoOnly")
	}
	oauth := Auth{Type: authtype.OAuth2, OAuth2: DefaultOAuth2Config()}
	if oauth.HasXenonOnlyOAuth2Fields() {
		t.Error("defaults are not Xenon-only settings")
	}
	oauth.OAuth2.AutoRefresh = false
	if !oauth.HasXenonOnlyOAuth2Fields() || (Auth{Type: authtype.Bearer}).HasXenonOnlyOAuth2Fields() {
		t.Error("HasXenonOnlyOAuth2Fields")
	}
	if len((Auth{Type: authtype.None}).CredentialFields()) != 0 {
		t.Error("no auth has no credentials")
	}
	if (NewEnvironment("e", "dev")).HasColor() || !(Environment{Color: envcolor.Cyan}).HasColor() || (Environment{}).HasColor() {
		t.Error("HasColor")
	}
}

func TestExportCountsAddEveryCount(t *testing.T) {
	one := ExportCounts{Collections: 1, Folders: 2, Requests: 3, Examples: 4, Environments: 5, Variables: 6}
	if got := one.Add(one); got != (ExportCounts{Collections: 2, Folders: 4, Requests: 6, Examples: 8, Environments: 10, Variables: 12}) {
		t.Errorf("add = %+v", got)
	}
}

func TestNewExportOptionsLooksAtEveryScope(t *testing.T) {
	secret := NewVariableRow("token", "")
	secret.Secret = true
	plain := NewVariableRow("base", "https://x")
	local := NewLocalVariables().
		WithCurrentValues(NewEnvironmentScopeRef("e"), map[string]string{"base": "http://local", "token": "t"}).
		WithCurrentValues(NewGlobalsScopeRef(), map[string]string{"token": "only-secret", "unknown": "x"})
	cases := []struct {
		scopes  []ExportScope
		secrets bool
		current bool
	}{
		{scopes: nil},
		{scopes: []ExportScope{NewExportScope(NewGlobalsScopeRef(), []VariableRow{plain})}},
		{scopes: []ExportScope{NewExportScope(NewGlobalsScopeRef(), []VariableRow{secret}), NewExportScope(NewEnvironmentScopeRef("e"), []VariableRow{plain})}, secrets: true, current: true},
		{scopes: []ExportScope{NewExportScope(NewEnvironmentScopeRef("e"), []VariableRow{plain}), NewExportScope(NewGlobalsScopeRef(), []VariableRow{secret})}, secrets: true, current: true},
		{scopes: []ExportScope{NewExportScope(NewEnvironmentScopeRef("e"), []VariableRow{secret})}, secrets: true},
		{scopes: []ExportScope{NewExportScope(NewEnvironmentScopeRef("other"), []VariableRow{plain})}},
	}
	for index, tc := range cases {
		options := NewExportOptions(exportitemkind.Environment, "dev", tc.scopes, local)
		if options.HasSecrets != tc.secrets || options.HasCurrentValues != tc.current || options.SuggestedFileName != "dev" || len(options.Formats) != 2 {
			t.Errorf("case %d: %+v", index, options)
		}
	}
}

func TestVariableRowWithCurrentAsInitial(t *testing.T) {
	secret := NewVariableRow("token", "")
	secret.Secret = true
	cases := []struct {
		row  VariableRow
		want string
	}{
		{NewVariableRow("base", "https://x"), "http://local"},
		{NewVariableRow("tenant", "acme"), "acme"},
		{secret, ""},
	}
	current := map[string]string{"base": "http://local", "tenant": "", "token": "t"}
	for _, tc := range cases {
		if got := tc.row.WithCurrentAsInitial(current).InitialValue; got != tc.want {
			t.Errorf("%s = %q, want %q", tc.row.Key, got, tc.want)
		}
	}
}
