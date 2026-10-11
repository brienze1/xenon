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
	request := ExportRequest{Format: exportformat.Postman, ItemKind: exportitemkind.Workspace, Shape: exportshape.Zip, Passphrase: "secret-pass"}
	if !request.IsPostman() || !request.IsArchive() {
		t.Error("expected a Postman workspace zip export")
	}
	if (ExportRequest{Shape: exportshape.Bundle}).IsArchive() {
		t.Error("a bundle is not an archive")
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

func TestExportRequestExportedScopes(t *testing.T) {
	scopes := []ExportScope{
		NewExportScope(NewEnvironmentScopeRef("env"), nil),
		NewExportScope(NewCollectionScopeRef("col"), nil),
		NewExportScope(NewGlobalsScopeRef(), nil),
	}
	cases := []struct {
		request ExportRequest
		want    int
	}{
		{ExportRequest{Format: exportformat.Postman, ItemKind: exportitemkind.Folder}, 3},
		{ExportRequest{Format: exportformat.XenonYaml, ItemKind: exportitemkind.Workspace}, 3},
		{ExportRequest{Format: exportformat.XenonYaml, ItemKind: exportitemkind.Collection}, 3},
		{ExportRequest{Format: exportformat.XenonYaml, ItemKind: exportitemkind.Folder}, 2},
		{ExportRequest{Format: exportformat.XenonYaml, ItemKind: exportitemkind.Request}, 2},
	}
	for _, tc := range cases {
		exported := tc.request.ExportedScopes(scopes)
		if len(exported) != tc.want {
			t.Errorf("%s/%s: %d scopes, want %d", tc.request.Format, tc.request.ItemKind, len(exported), tc.want)
		}
		for _, scope := range exported {
			if tc.want == 2 && scope.Ref.IsCollection() {
				t.Errorf("%s/%s kept the collection scope", tc.request.Format, tc.request.ItemKind)
			}
		}
	}
}

func TestExportScopeCurrentValuesAsInitial(t *testing.T) {
	scope := NewExportScope(NewEnvironmentScopeRef("env"), []VariableRow{
		NewVariableRow("url", "https://a"),
		{Key: "token", Secret: true, Enabled: true},
	})
	local := NewLocalVariables().WithCurrentValues(scope.Ref, map[string]string{"token": "current-secret"})
	if rows, changed := scope.CurrentValuesAsInitial(local); changed || rows[0].InitialValue != "https://a" || rows[1].InitialValue != "" {
		t.Errorf("unchanged rows = %+v, changed = %v", rows, changed)
	}
	local = local.WithCurrentValues(scope.Ref, map[string]string{"token": "current-secret", "url": "http://localhost"})
	rows, changed := scope.CurrentValuesAsInitial(local)
	if !changed || rows[0].InitialValue != "http://localhost" || rows[1].InitialValue != "" || len(rows) != 2 {
		t.Errorf("changed rows = %+v, changed = %v", rows, changed)
	}
}

func TestExportScopeSecretValues(t *testing.T) {
	scope := NewExportScope(NewEnvironmentScopeRef("env"), []VariableRow{
		NewVariableRow("url", "https://a"),
		{Key: "token", Secret: true},
		{Key: "unset", Secret: true},
	})
	vault := NewVault("local", "", "", map[string]map[string]string{"env:env": {"token": "t-1", "url": "not-secret", "unset": ""}}).Unlocked()
	if values := scope.SecretValues(vault); !reflect.DeepEqual(values, map[string]string{"token": "t-1"}) {
		t.Errorf("secret values = %v", values)
	}
	if values := scope.SecretValues(NewUnopenedVault("local")); len(values) != 0 {
		t.Errorf("an unreadable vault gave %v", values)
	}
}

func TestSecretBundle(t *testing.T) {
	bundle := NewSecretBundle(map[string]map[string]string{"env:a": {"k": "v"}, "collection:b": {}, "globals": {"g": "1"}})
	if bundle.IsEmpty() || len(bundle.Scopes) != 2 {
		t.Fatalf("bundle = %+v", bundle.Scopes)
	}
	if got := bundle.For(NewEnvironmentScopeRef("a")); !reflect.DeepEqual(got, map[string]string{"k": "v"}) {
		t.Errorf("For(env:a) = %v", got)
	}
	if got := bundle.For(NewCollectionScopeRef("b")); len(got) != 0 {
		t.Errorf("For(collection:b) = %v", got)
	}
	if !NewSecretBundle(nil).IsEmpty() {
		t.Error("a bundle without scopes is empty")
	}
	logged, _ := bundle.LogValue().(map[string]any)
	if counts, _ := logged["scopes"].(map[string]int); counts["env:a"] != 1 || counts["globals"] != 1 || len(counts) != 2 {
		t.Errorf("log value = %v", logged)
	}
}

func TestXenonExport(t *testing.T) {
	request := ExportRequest{ItemKind: exportitemkind.Environment, ItemID: "env", Shape: exportshape.File, CurrentValuesAsInitial: true}
	export := NewXenonExport(request)
	if export.Kind != exportitemkind.Environment || export.ItemID != "env" || !export.CurrentValuesAsInitial || export.IsArchive() || export.HasVariables() || export.HasSecrets() {
		t.Fatalf("export = %+v", export)
	}
	rows := []VariableRow{NewVariableRow("url", "http://localhost")}
	changed := export.WithVariables(NewEnvironmentScopeRef("env"), rows).WithSecrets("armored").WithNode(NewRequestExportNode(NewRequest("r", "R"), "R"))
	if found, ok := changed.VariablesFor(NewEnvironmentScopeRef("env")); !ok || !reflect.DeepEqual(found, rows) || !changed.HasVariables() {
		t.Errorf("variables = %v, %v", found, ok)
	}
	if _, ok := changed.VariablesFor(NewGlobalsScopeRef()); ok {
		t.Error("globals have no variables")
	}
	if export.HasVariables() || !changed.HasSecrets() || changed.Node == nil || changed.Node.Name() != "R" {
		t.Errorf("changed = %+v, original = %+v", changed, export)
	}
	logged, _ := changed.LogValue().(map[string]any)
	if logged["hasSecrets"] != true || logged["secrets"] != nil {
		t.Errorf("log value = %v", logged)
	}
	if !(XenonExport{Shape: exportshape.Zip}).IsArchive() {
		t.Error("a zip is an archive")
	}
}

func TestExportProgressMilestones(t *testing.T) {
	cases := []struct {
		progress ExportProgress
		expected bool
	}{
		{ExportProgress{Done: 50, Total: 201}, true},
		{ExportProgress{Done: 200, Total: 1007}, true},
		{ExportProgress{Done: 50, Total: 200}, false},
		{ExportProgress{Done: 0, Total: 1007}, false},
		{ExportProgress{Done: 51, Total: 1007}, false},
		{ExportProgress{Done: 1000, Total: 1000}, false},
		{ExportProgress{Done: 1000, Total: 1001}, true},
	}
	for _, c := range cases {
		if got := c.progress.IsMilestone(); got != c.expected {
			t.Errorf("%+v: IsMilestone() = %v, expected %v", c.progress, got, c.expected)
		}
	}
}

func TestExportFileNameForShape(t *testing.T) {
	cases := []struct {
		name     string
		format   enums.ExportFormat
		shape    enums.ExportShape
		expected string
	}{
		{"shop-api.zip", exportformat.XenonYaml, exportshape.Bundle, "shop-api.yaml"},
		{"shop-api.yaml", exportformat.XenonYaml, exportshape.Zip, "shop-api.zip"},
		{"shop-api", exportformat.XenonYaml, exportshape.Bundle, "shop-api.yaml"},
		{"local.yaml", exportformat.XenonYaml, exportshape.File, "local.yaml"},
		{"local.yaml", exportformat.XenonYaml, exportshape.Unknown, "local.yaml"},
		{"shop.postman_collection.json", exportformat.Postman, exportshape.Bundle, "shop.postman_collection.json"},
	}
	for _, c := range cases {
		if got := ExportFileNameForShape(c.name, c.format, c.shape); got != c.expected {
			t.Errorf("%s %s %s: got %q, expected %q", c.name, c.format, c.shape, got, c.expected)
		}
	}
}

func TestExportPickerTitle(t *testing.T) {
	if got := ExportPickerTitle(exportformat.Postman); got != "Export as Postman" {
		t.Errorf("postman title = %q", got)
	}
	if got := ExportPickerTitle(exportformat.XenonYaml); got != "Export as Xenon YAML" {
		t.Errorf("xenon title = %q", got)
	}
}

func TestExportNodeIdentityAuthHeadersAndScope(t *testing.T) {
	collection := NewCollection("c", "Billing")
	collection.Auth = Auth{Type: authtype.Bearer}
	request := NewRequest("r", "Create")
	request.Auth = Auth{Type: authtype.Basic}
	container := NewContainerExportNode(collection, "Billing", nil)
	leaf := NewRequestExportNode(request, "Billing/Create")
	if container.ID() != "c" || leaf.ID() != "r" {
		t.Errorf("ids = %q, %q", container.ID(), leaf.ID())
	}
	if container.Auth().Type != authtype.Bearer || leaf.Auth().Type != authtype.Basic {
		t.Errorf("auth = %v, %v", container.Auth().Type, leaf.Auth().Type)
	}
	if container.HasContainerHeaders() || leaf.HasContainerHeaders() {
		t.Error("no headers yet")
	}
	collection.Headers = []KeyValue{{Key: "X", Value: "1", Enabled: true}}
	if !NewContainerExportNode(collection, "Billing", nil).HasContainerHeaders() {
		t.Error("a container with headers has container headers")
	}
	folder := NewContainerExportNode(NewFolder("f", "Invoices"), "Billing/Invoices", nil)
	if folder.ScopeRefWithin(nil) != NewCollectionScopeRef("f") {
		t.Errorf("own scope = %v", folder.ScopeRefWithin(nil))
	}
	if folder.ScopeRefWithin([]Collection{collection, NewFolder("g", "G")}) != NewCollectionScopeRef("c") {
		t.Errorf("root scope = %v", folder.ScopeRefWithin([]Collection{collection}))
	}
}

func TestExportPathAndExportableChildren(t *testing.T) {
	if got := ExportPath([]Collection{NewCollection("c", "Billing"), NewFolder("f", "Invoices")}, "Create"); got != "Billing/Invoices/Create" {
		t.Errorf("path = %q", got)
	}
	if got := ExportPath(nil, "Billing"); got != "Billing" {
		t.Errorf("root path = %q", got)
	}
	kept := NewRequestNode(NewRequest("r", "R"), "r.yaml")
	ignored := NewRequestNode(NewRequest("i", "I"), "i.yaml")
	ignored.Ignored = true
	broken := NewErrorNode("x", "x.yaml", "bad", false)
	parent := TreeNode{Children: []TreeNode{kept, ignored, broken}}
	children := ExportableChildren(parent)
	if len(children) != 1 || children[0].ID != "r" {
		t.Errorf("children = %+v", children)
	}
}

func TestEnvironmentExportCounts(t *testing.T) {
	environment := NewEnvironment("e", "dev")
	environment.Variables = []VariableRow{NewVariableRow("a", "1"), NewVariableRow("b", "2")}
	if got := environment.ExportCounts(); got != (ExportCounts{Environments: 1, Variables: 2}) {
		t.Errorf("counts = %+v", got)
	}
	if got := NewVariablesExportCounts(environment.Variables); got != (ExportCounts{Variables: 2}) {
		t.Errorf("variables counts = %+v", got)
	}
	malformed := NewMalformedEnvironment("bad.yaml", "broken")
	wellFormed := WellFormedEnvironments([]Environment{malformed, environment})
	if len(wellFormed) != 1 || wellFormed[0].ID != "e" {
		t.Errorf("well formed = %+v", wellFormed)
	}
}

func TestXenonExportKindsAndSecrets(t *testing.T) {
	workspace := XenonExport{Kind: exportitemkind.Workspace}
	environment := XenonExport{Kind: exportitemkind.Environment}
	if !workspace.IsWorkspace() || workspace.IsEnvironment() || !environment.IsEnvironment() || environment.IsWorkspace() {
		t.Error("kinds are told apart")
	}
	if workspace.ExportedSecrets() != exportsecrets.Omitted || workspace.WithSecrets("armored").ExportedSecrets() != exportsecrets.Included {
		t.Error("secrets are included only when the export carries them")
	}
}
