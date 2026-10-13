package model

import (
	"strings"
	"testing"
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

var exportedAt = time.Date(2026, 10, 11, 7, 0, 0, 0, time.FixedZone("BRT", -3*60*60))

func encodeDocument(t *testing.T, document *yaml.Node) string {
	t.Helper()
	data, err := yamlnode.Encode(document)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestExportManifestDocument(t *testing.T) {
	with := NewExportManifest("workspace", exportedAt, "Xenon/1.0.0", true, true)
	if !with.HasSecretsFile() {
		t.Error("a manifest with secrets points at the secrets file")
	}
	want := "xenon: 1\nkind: export\nroot: workspace\nexportedAt: \"2026-10-11T10:00:00Z\"\nexportedBy: Xenon/1.0.0\ncurrentValuesAsInitial: true\nsecrets: file\n"
	if got := encodeDocument(t, with.Document()); got != want {
		t.Errorf("manifest = %q", got)
	}
	without := NewExportManifest("folder", exportedAt, "b", false, false)
	if without.HasSecretsFile() || !strings.HasSuffix(encodeDocument(t, without.Document()), "currentValuesAsInitial: false\nsecrets: null\n") {
		t.Errorf("manifest = %q", encodeDocument(t, without.Document()))
	}
}

func TestExportBundleDocument(t *testing.T) {
	item := WithChildren(yamlnode.MappingNode(yamlnode.String("kind", "collection")), []*yaml.Node{
		WithChildren(yamlnode.MappingNode(yamlnode.String("kind", "folder")), nil),
	})
	bundle := NewExportBundle("collection", exportedAt, "b", item, "line1\nline2\n")
	want := "xenon: 1\nkind: bundle\nroot: collection\nexportedAt: \"2026-10-11T10:00:00Z\"\nexportedBy: b\nitem:\n  kind: collection\n  children:\n    - kind: folder\n      children: []\nsecrets: |\n  line1\n  line2\n"
	if got := encodeDocument(t, bundle.Document()); got != want {
		t.Errorf("bundle = %q", got)
	}
	plain := NewExportBundle("folder", exportedAt, "b", yamlnode.MappingNode(yamlnode.String("kind", "folder")), "")
	if got := encodeDocument(t, plain.Document()); !strings.HasSuffix(got, "secrets: null\n") {
		t.Errorf("bundle = %q", got)
	}
}

func TestExportSecretsVersions(t *testing.T) {
	bundle := entity.NewSecretBundle(map[string]map[string]string{"globals": {"k": "v"}})
	stored := NewExportSecrets(bundle)
	if stored.Version != 1 || stored.IsNewer() || stored.ToEntity().For(entity.NewGlobalsScopeRef())["k"] != "v" {
		t.Errorf("stored = %+v", stored)
	}
	if !(&ExportSecrets{Version: 2}).IsNewer() || (&ExportSecrets{Version: 0}).IsNewer() {
		t.Error("only a version above 1 is newer")
	}
}

func TestXenonHeaderVariablesScope(t *testing.T) {
	cases := []struct {
		header XenonHeader
		scope  string
		owned  bool
	}{
		{XenonHeader{Kind: "environment", Id: "e"}, "env:e", true},
		{XenonHeader{Kind: "Environment", Id: ""}, "", false},
		{XenonHeader{Kind: "collection", Id: "c"}, "collection:c", true},
		{XenonHeader{Kind: "collection"}, "", false},
		{XenonHeader{Kind: "workspace"}, "globals", true},
		{XenonHeader{Kind: "folder", Id: "f"}, "", false},
		{XenonHeader{Kind: "request", Id: "r"}, "", false},
	}
	for _, tc := range cases {
		scope, owned := tc.header.VariablesScope()
		if owned != tc.owned || (owned && scope.String() != tc.scope) {
			t.Errorf("%+v: %s %v", tc.header, scope.String(), owned)
		}
	}
	probe := (&XenonHeader{Xenon: 2, Kind: "export", Root: "workspace"}).Probed(entity.ImportProbe{Name: "x"})
	if probe.Name != "x" || probe.XenonFormat != 2 || probe.XenonKind != "export" || probe.XenonRoot != "workspace" {
		t.Errorf("probe = %+v", probe)
	}
}

func TestVariablesField(t *testing.T) {
	field := VariablesField([]entity.VariableRow{entity.NewVariableRow("a", "b")})
	if field.Key != "variables" || len(field.Value.Content) != 1 {
		t.Errorf("field = %+v", field)
	}
}
