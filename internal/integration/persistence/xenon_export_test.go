package persistence

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/agecrypt"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/timeutils"
)

const (
	exportCollectionId = "01JX0000000000000000000001"
	exportFolderId     = "01JX0000000000000000000002"
	exportRequestId    = "01JX0000000000000000000011"
	exportSecondId     = "01JX0000000000000000000012"
	exportEnvId        = "01JX0000000000000000000101"
	exportCollection   = "xenon: 1\nkind: collection\nid: " + exportCollectionId + "\nname: Shop\norder:\n  - products\n  - b.yaml\nvariables:\n  - key: page\n    value: \"20\"\n  - key: token\n    secret: true\n"
	exportFolder       = "xenon: 1\nkind: folder\nid: " + exportFolderId + "\nname: Products\n"
	exportRequest      = "xenon: 1\nkind: request\nid: " + exportRequestId + "\nname: A\nmethod: GET\nurl: https://a.test\n"
	exportSecond       = "xenon: 1\nkind: request\nid: " + exportSecondId + "\nname: B\nmethod: POST\nurl: https://b.test"
	exportEnvironment  = "xenon: 1\nkind: environment\nid: " + exportEnvId + "\nname: local\nvariables:\n  - key: url\n    value: https://local\n  - key: key\n    secret: true\n"
	exportWorkspace    = "xenon: 1\nkind: workspace\nid: 01JX0000000000000000000100\nname: Acme\nvariables:\n  - key: region\n    value: eu\n"
)

func writeExportWorkspace(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "acme")
	files := map[string]string{
		"workspace.yaml":                         exportWorkspace,
		".gitignore":                             ".xenon/\n",
		"environments/local.yaml":                exportEnvironment,
		"collections/shop/collection.yaml":       exportCollection,
		"collections/shop/b.yaml":                exportSecond,
		"collections/shop/products/folder.yaml":  exportFolder,
		"collections/shop/products/a.yaml":       exportRequest,
		"collections/shop/.a.yaml.xenon-tmp":     "half",
		"files/payload.bin":                      "\x00\x01",
		"files/.git":                             "gitdir: elsewhere\n",
		".xenon/state.yaml":                      "local: only\n",
		".xenon-import-1/0/x.yaml":               "staged: true\n",
		".git/HEAD":                              "ref: main\n",
		"collections/shop/products/notes.txt":    "kept",
		"collections/shop/products/.hidden.yaml": "kept: true\n",
	}
	for name, content := range files {
		writeTestFile(t, filepath.Join(root, filepath.FromSlash(name)), content)
	}
	return root
}

func fixedExportTime(t *testing.T) {
	t.Helper()
	timeutils.SetCurrentTimeProvider(func() time.Time { return time.Date(2026, 10, 11, 12, 0, 0, 0, time.FixedZone("x", 3600)) })
	t.Cleanup(func() { timeutils.SetCurrentTimeProvider(time.Now) })
}

func entriesByPath(entries []entity.ExportEntry) map[string]string {
	byPath := make(map[string]string, len(entries))
	for _, entry := range entries {
		byPath[entry.Path] = string(entry.Data)
	}
	return byPath
}

func sortedPaths(byPath map[string]string) string {
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return strings.Join(paths, ",")
}

func expectErrorCode(t *testing.T, err error, code string) {
	t.Helper()
	custom, isCustom := err.(errs.BaseError)
	if !isCustom || custom.Code() != code {
		t.Fatalf("error = %v, want %s", err, code)
	}
}

func TestWriteXenonExportWorkspaceZip(t *testing.T) {
	fixedExportTime(t)
	root := writeExportWorkspace(t)
	export := entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Workspace, Shape: exportshape.Zip, CurrentValuesAsInitial: true}).
		WithVariables(entity.NewEnvironmentScopeRef(exportEnvId), []entity.VariableRow{entity.NewVariableRow("url", "http://current"), {Key: "key", Secret: true, Enabled: true}}).
		WithVariables(entity.NewGlobalsScopeRef(), []entity.VariableRow{entity.NewVariableRow("region", "us")}).
		WithSecrets("-----BEGIN AGE ENCRYPTED FILE-----\nabc\n-----END AGE ENCRYPTED FILE-----\n")
	entries, err := NewXenonExportPersistence(NewWorkspaceIndexCache()).WriteXenonExport(context.Background(), root, export)
	if err != nil {
		t.Fatal(err)
	}
	byPath := entriesByPath(entries)
	want := "acme/.gitignore,acme/collections/shop/b.yaml,acme/collections/shop/collection.yaml,acme/collections/shop/products/.hidden.yaml," +
		"acme/collections/shop/products/a.yaml,acme/collections/shop/products/folder.yaml,acme/collections/shop/products/notes.txt," +
		"acme/environments/local.yaml,acme/files/.git,acme/files/payload.bin,acme/workspace.yaml,xenon-export.yaml,xenon-secrets.age"
	if got := sortedPaths(byPath); got != want {
		t.Fatalf("paths = %s", got)
	}
	if entries[0].Path != exportManifestFile {
		t.Errorf("the manifest comes first, got %s", entries[0].Path)
	}
	manifest := byPath[exportManifestFile]
	for _, line := range []string{"xenon: 1\n", "kind: export\n", "root: workspace\n", "exportedAt: \"2026-10-11T11:00:00Z\"\n", "exportedBy: Xenon/", "currentValuesAsInitial: true\n", "secrets: file\n"} {
		if !strings.Contains(manifest, line) {
			t.Errorf("manifest lacks %q:\n%s", line, manifest)
		}
	}
	if byPath[exportSecretsFile] != export.Secrets {
		t.Errorf("secrets file = %q", byPath[exportSecretsFile])
	}
	if env := byPath["acme/environments/local.yaml"]; !strings.Contains(env, "value: http://current") || strings.Contains(env, "https://local") {
		t.Errorf("environment = %s", env)
	}
	if globals := byPath["acme/workspace.yaml"]; !strings.Contains(globals, "value: us\n") {
		t.Errorf("workspace = %s", globals)
	}
	for _, unchanged := range []string{"collections/shop/collection.yaml", "collections/shop/b.yaml", "files/payload.bin", ".gitignore"} {
		original, _ := os.ReadFile(filepath.Join(root, filepath.FromSlash(unchanged)))
		if byPath["acme/"+unchanged] != string(original) {
			t.Errorf("%s changed: %q", unchanged, byPath["acme/"+unchanged])
		}
	}
}

func TestWriteXenonExportWithoutOptions(t *testing.T) {
	root := writeExportWorkspace(t)
	export := entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Workspace, Shape: exportshape.Zip})
	entries, err := NewXenonExportPersistence(NewWorkspaceIndexCache()).WriteXenonExport(context.Background(), root, export)
	if err != nil {
		t.Fatal(err)
	}
	byPath := entriesByPath(entries)
	if _, found := byPath[exportSecretsFile]; found {
		t.Error("no secrets file without secrets")
	}
	if manifest := byPath[exportManifestFile]; !strings.Contains(manifest, "secrets: null\n") || !strings.Contains(manifest, "currentValuesAsInitial: false\n") {
		t.Errorf("manifest = %s", manifest)
	}
	if byPath["acme/environments/local.yaml"] != exportEnvironment || byPath["acme/workspace.yaml"] != exportWorkspace {
		t.Error("variable files must be exported as they are")
	}
}

func TestWriteXenonExportEnvironmentFile(t *testing.T) {
	root := writeExportWorkspace(t)
	persistence := NewXenonExportPersistence(NewWorkspaceIndexCache())
	export := entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Environment, ItemID: exportEnvId, Shape: exportshape.File}).WithSecrets("line1\nline2\n")
	entries, err := persistence.WriteXenonExport(context.Background(), root, export)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Path != "local.yaml" || string(entries[0].Data) != exportEnvironment+"secrets: |\n  line1\n  line2\n" {
		t.Fatalf("entries = %+v", entries)
	}
	export = entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Environment, ItemID: "missing", Shape: exportshape.File})
	if _, err := persistence.WriteXenonExport(context.Background(), root, export); err == nil {
		t.Error("an unknown environment must fail")
	}
}

func TestWriteXenonExportRequestFileWithoutTrailingNewline(t *testing.T) {
	root := writeExportWorkspace(t)
	persistence := NewXenonExportPersistence(NewWorkspaceIndexCache())
	export := entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Request, ItemID: exportSecondId, Shape: exportshape.File})
	entries, err := persistence.WriteXenonExport(context.Background(), root, export)
	if err != nil || len(entries) != 1 || entries[0].Path != "b.yaml" || string(entries[0].Data) != exportSecond {
		t.Fatalf("entries = %+v, %v", entries, err)
	}
	entries, err = persistence.WriteXenonExport(context.Background(), root, export.WithSecrets("s\n"))
	if err != nil || string(entries[0].Data) != exportSecond+"\nsecrets: |\n  s\n" {
		t.Fatalf("entries = %q, %v", entries[0].Data, err)
	}
	if data := withSecretsBlock(nil, "s"); string(data) != "secrets: |\n  s\n" {
		t.Errorf("empty data = %q", data)
	}
	export = entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Request, ItemID: "01JX0000000000000000000999", Shape: exportshape.File})
	_, err = persistence.WriteXenonExport(context.Background(), root, export)
	expectErrorCode(t, err, "EXP-16404")
}

func TestWriteXenonExportFolderZip(t *testing.T) {
	root := writeExportWorkspace(t)
	export := entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Folder, ItemID: exportFolderId, Shape: exportshape.Zip})
	entries, err := NewXenonExportPersistence(NewWorkspaceIndexCache()).WriteXenonExport(context.Background(), root, export)
	if err != nil {
		t.Fatal(err)
	}
	byPath := entriesByPath(entries)
	if got := sortedPaths(byPath); got != "products/.hidden.yaml,products/a.yaml,products/folder.yaml,products/notes.txt,xenon-export.yaml" {
		t.Fatalf("paths = %s", got)
	}
	if !strings.Contains(byPath[exportManifestFile], "root: folder\n") || byPath["products/a.yaml"] != exportRequest {
		t.Errorf("manifest = %s", byPath[exportManifestFile])
	}
}

func TestWriteXenonExportCollectionBundle(t *testing.T) {
	fixedExportTime(t)
	root := writeExportWorkspace(t)
	request := func(id, name string) entity.ExportNode {
		return entity.NewRequestExportNode(entity.NewRequest(id, name), name)
	}
	folder := entity.NewContainerExportNode(entity.Collection{ID: exportFolderId, Kind: itemkind.Folder, Name: "Products"}, "Shop/Products", []entity.ExportNode{request(exportRequestId, "A")})
	empty := entity.NewContainerExportNode(entity.Collection{ID: exportFolderId, Kind: itemkind.Folder, Name: "Products"}, "Shop/Products", []entity.ExportNode{})
	node := entity.NewContainerExportNode(entity.Collection{ID: exportCollectionId, Kind: itemkind.Collection, Name: "Shop"}, "Shop", []entity.ExportNode{request(exportSecondId, "B"), folder, empty})
	export := entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Collection, ItemID: exportCollectionId, Shape: exportshape.Bundle}).
		WithNode(node).
		WithVariables(entity.NewCollectionScopeRef(exportCollectionId), []entity.VariableRow{entity.NewVariableRow("page", "50"), {Key: "token", Secret: true, Enabled: true}}).
		WithSecrets("armored\n")
	persistence := NewXenonExportPersistence(NewWorkspaceIndexCache())
	entries, err := persistence.WriteXenonExport(context.Background(), root, export)
	if err != nil {
		t.Fatal(err)
	}
	bundle := string(entries[0].Data)
	for _, part := range []string{
		"xenon: 1\nkind: bundle\nroot: collection\nexportedAt: \"2026-10-11T11:00:00Z\"\nexportedBy: Xenon/",
		"item:\n  xenon: 1\n  kind: collection\n  id: " + exportCollectionId,
		"value: \"50\"",
		"  children:\n    - xenon: 1\n      kind: request\n      id: " + exportSecondId,
		"    - xenon: 1\n      kind: folder\n      id: " + exportFolderId,
		"      children:\n        - xenon: 1\n          kind: request\n          id: " + exportRequestId,
		"      children: []\n",
		"secrets: |\n  armored\n",
	} {
		if !strings.Contains(bundle, part) {
			t.Errorf("bundle lacks %q:\n%s", part, bundle)
		}
	}
	if len(entries) != 1 || entries[0].Path != "Shop" {
		t.Errorf("entries = %d, path %s", len(entries), entries[0].Path)
	}
	plain, err := persistence.WriteXenonExport(context.Background(), root, entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Collection, ItemID: exportCollectionId, Shape: exportshape.Bundle}).WithNode(node))
	if err != nil || !strings.HasSuffix(string(plain[0].Data), "secrets: null\n") || !strings.Contains(string(plain[0].Data), "value: \"20\"") {
		t.Errorf("plain bundle = %s, %v", plain[0].Data, err)
	}
}

func TestWriteXenonExportBundleFailures(t *testing.T) {
	root := writeExportWorkspace(t)
	persistence := NewXenonExportPersistence(NewWorkspaceIndexCache())
	node := entity.NewContainerExportNode(entity.Collection{ID: exportCollectionId, Kind: itemkind.Collection}, "Shop", []entity.ExportNode{
		entity.NewRequestExportNode(entity.NewRequest(exportSecondId, "B"), "B"),
	})
	export := entity.NewXenonExport(entity.ExportRequest{ItemKind: exportitemkind.Collection, ItemID: exportCollectionId, Shape: exportshape.Bundle}).WithNode(node)
	if _, err := persistence.WriteXenonExport(context.Background(), root, export); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "collections/shop/b.yaml"), "name: [x\n")
	_, err := persistence.WriteXenonExport(context.Background(), root, export)
	expectErrorCode(t, err, "EXP-22422")
	if err := os.Remove(filepath.Join(root, "collections/shop/b.yaml")); err != nil {
		t.Fatal(err)
	}
	_, err = bundleEntries(&entity.TreeIndex{}, export)
	expectErrorCode(t, err, "EXP-21500")
}

func TestWithExportedVariablesLeavesOtherFilesAlone(t *testing.T) {
	export := entity.NewXenonExport(entity.ExportRequest{}).WithVariables(entity.NewEnvironmentScopeRef("e"), []entity.VariableRow{entity.NewVariableRow("a", "b")})
	cases := map[string]string{
		"x.txt":  "kind: environment\nid: e\n",
		"y.yaml": "kind: [\n",
		"z.yaml": "- not a mapping\n",
		"w.yaml": "xenon: one\nkind: environment\nid: e\n",
		"v.yaml": "xenon: 1\nkind: environment\nid: other\n",
		"u.yaml": "xenon: 1\nkind: request\nid: e\n",
	}
	for path, data := range cases {
		if got := withExportedVariables(path, []byte(data), export); string(got) != data {
			t.Errorf("%s changed to %q", path, got)
		}
	}
	if got := withExportedVariables("e.YAML", []byte("xenon: 1\nkind: environment\nid: e\n"), export); !strings.Contains(string(got), "variables:\n  - key: a\n    value: b\n") {
		t.Errorf("environment = %q", got)
	}
	if got := withExportedVariables("e.yaml", []byte("xenon: 1\nkind: environment\nid: e\n"), entity.NewXenonExport(entity.ExportRequest{})); string(got) != "xenon: 1\nkind: environment\nid: e\n" {
		t.Errorf("without variables = %q", got)
	}
}

func TestExportedFilesReportsUnreadableFolders(t *testing.T) {
	_, err := exportedFiles(filepath.Join(t.TempDir(), "missing"), "x", entity.XenonExport{})
	expectErrorCode(t, err, "EXP-18500")
	_, err = fileEntries(filepath.Join(t.TempDir(), "missing.yaml"), entity.XenonExport{})
	expectErrorCode(t, err, "EXP-19500")
}

func TestIsLocalOnlyDir(t *testing.T) {
	for name, want := range map[string]bool{".xenon": true, ".git": true, ".xenon-import-01": true, ".xenonx": false, "collections": false, ".github": false} {
		if got := isLocalOnlyDir(name); got != want {
			t.Errorf("%s: %v", name, got)
		}
	}
}

func TestExportSecretsRoundTrip(t *testing.T) {
	t.Setenv("XENON_SCRYPT_LOG_N", "10")
	persistence := NewXenonExportPersistence(NewWorkspaceIndexCache())
	bundle := entity.NewSecretBundle(map[string]map[string]string{"env:" + exportEnvId: {"key": "s3cr3t"}})
	armored, err := persistence.Encrypt(context.Background(), bundle, "correct horse")
	if err != nil || !strings.HasPrefix(armored, "-----BEGIN AGE ENCRYPTED FILE-----\n") || strings.Contains(armored, "s3cr3t") {
		t.Fatalf("armored = %q, %v", armored, err)
	}
	decrypted, err := persistence.Decrypt(context.Background(), armored, "correct horse")
	if err != nil || decrypted.For(entity.NewEnvironmentScopeRef(exportEnvId))["key"] != "s3cr3t" {
		t.Fatalf("decrypted = %+v, %v", decrypted, err)
	}
	_, err = persistence.Decrypt(context.Background(), armored, "wrong horse")
	expectErrorCode(t, err, "IMP-08401")
	_, err = persistence.Encrypt(context.Background(), bundle, "")
	expectErrorCode(t, err, "EXP-15500")
	for plaintext, code := range map[string]string{"not json": "IMP-44422", `{"version": 2, "secrets": {}}`: "IMP-45422"} {
		sealed, _ := agecrypt.EncryptArmoredWithPassphrase([]byte(plaintext), "correct horse", 10)
		_, err = persistence.Decrypt(context.Background(), string(sealed), "correct horse")
		expectErrorCode(t, err, code)
	}
	sealed, _ := agecrypt.EncryptArmoredWithPassphrase([]byte(`{"version": 1, "secrets": {"globals": {"g": "1"}}}`), "correct horse", 10)
	if decrypted, err := persistence.Decrypt(context.Background(), string(sealed), "correct horse"); err != nil || decrypted.For(entity.NewGlobalsScopeRef())["g"] != "1" {
		t.Errorf("version 1 = %+v, %v", decrypted, err)
	}
}

func TestReadXenonExportRoundTrips(t *testing.T) {
	root := writeExportWorkspace(t)
	ctx := context.Background()
	persistence := NewXenonExportPersistence(NewWorkspaceIndexCache())
	request, err := persistence.ReadXenonExport(ctx, entity.ImportInput{Kind: importkind.XenonRequest, Documents: []entity.ImportDocument{{Path: "a.yaml", Data: []byte(exportRequest + "secrets: |\n  armored\n")}}})
	if err != nil || request.Root == nil || request.Root.Request.ID != exportRequestId || request.Root.Request.URL != "https://a.test" || request.Secrets != "armored\n" {
		t.Fatalf("request = %+v, %v", request, err)
	}
	environment, err := persistence.ReadXenonExport(ctx, entity.ImportInput{Kind: importkind.XenonEnvironment, Documents: []entity.ImportDocument{{Path: "dir/local.yaml", Data: []byte(exportEnvironment)}}})
	if err != nil || environment.Environment == nil || environment.Environment.ID != exportEnvId || environment.Environment.FileName != "local.yaml" || len(environment.Environment.Variables) != 2 || environment.Secrets != "" {
		t.Fatalf("environment = %+v, %v", environment, err)
	}
	bundle := "xenon: 1\nkind: bundle\nroot: collection\nitem:\n  xenon: 1\n  kind: collection\n  id: c1\n  name: Shop\n  children:\n    - xenon: 1\n      kind: request\n      id: r1\n      name: A\n    - xenon: 1\n      kind: folder\n      id: f1\n      name: F\n      children: []\nsecrets: armored\n"
	read, err := persistence.ReadXenonExport(ctx, entity.ImportInput{Kind: importkind.XenonBundle, Documents: []entity.ImportDocument{{Path: "b.yaml", Data: []byte(bundle)}}})
	if err != nil || read.Root.Container.ID != "c1" || read.Root.Container.Kind != itemkind.Collection || len(read.Root.Children) != 2 || read.Root.Children[0].Request.ID != "r1" || read.Root.Children[1].Container.Kind != itemkind.Folder || read.Secrets != "armored" {
		t.Fatalf("bundle = %+v, %v", read, err)
	}
	zip := filepath.Join(t.TempDir(), "zip")
	writeTestFile(t, filepath.Join(zip, exportManifestFile), "xenon: 1\nkind: export\nroot: collection\n")
	writeTestFile(t, filepath.Join(zip, exportSecretsFile), "sealed")
	if err := os.CopyFS(filepath.Join(zip, "shop"), os.DirFS(filepath.Join(root, "collections", "shop"))); err != nil {
		t.Fatal(err)
	}
	input := entity.ImportInput{Kind: importkind.XenonZip, Root: zip, Probe: entity.ImportProbe{XenonRoot: "collection", Entries: []string{exportManifestFile, ".hidden/x", "shop/collection.yaml", "shop/products/a.yaml"}}}
	read, err = persistence.ReadXenonExport(ctx, input)
	if err != nil || read.Root.Container.ID != exportCollectionId || read.Secrets != "sealed" {
		t.Fatalf("zip = %+v, %v", read, err)
	}
	if names := read.Root.ChildNames(); strings.Join(names, ",") != "Products,B" || read.Root.Children[0].Children[0].Request.ID != exportRequestId {
		t.Errorf("zip children = %v", names)
	}
	workspace, err := persistence.ReadXenonExport(ctx, entity.ImportInput{Kind: importkind.XenonWorkspaceFolder, Root: root})
	if err != nil || workspace.Workspace == nil || workspace.Workspace.Name != "Acme" || workspace.Workspace.Folder != root || workspace.Secrets != "" {
		t.Fatalf("workspace = %+v, %v", workspace, err)
	}
	if counts := workspace.Workspace.Counts; counts.Collections != 1 || counts.Folders != 1 || counts.Requests != 2 || counts.Environments != 1 {
		t.Errorf("counts = %+v", counts)
	}
	zipped := filepath.Join(t.TempDir(), "wszip")
	writeTestFile(t, filepath.Join(zipped, exportSecretsFile), "sealed")
	if err := os.CopyFS(filepath.Join(zipped, "acme"), os.DirFS(root)); err != nil {
		t.Fatal(err)
	}
	workspace, err = persistence.ReadXenonExport(ctx, entity.ImportInput{Kind: importkind.XenonWorkspaceZip, Root: zipped, Probe: entity.ImportProbe{Entries: []string{exportManifestFile, "acme/workspace.yaml"}}})
	if err != nil || workspace.Workspace.Folder != filepath.Join(zipped, "acme") || workspace.Secrets != "sealed" {
		t.Fatalf("workspace zip = %+v, %v", workspace, err)
	}
}

func TestReadXenonExportFailures(t *testing.T) {
	ctx := context.Background()
	persistence := NewXenonExportPersistence(NewWorkspaceIndexCache())
	document := func(kind enums.ImportKind, data string) entity.ImportInput {
		return entity.ImportInput{Kind: kind, Documents: []entity.ImportDocument{{Path: "dir/x.yaml", Data: []byte(data)}}}
	}
	cases := map[string]struct {
		input entity.ImportInput
		code  string
	}{
		"request syntax":    {document(importkind.XenonRequest, "name: [\n"), "IMP-47422"},
		"request header":    {document(importkind.XenonRequest, "xenon: one\n"), "IMP-47422"},
		"request fields":    {document(importkind.XenonRequest, "xenon: 1\nheaders: x\n"), "IMP-47422"},
		"env syntax":        {document(importkind.XenonEnvironment, "- x\n"), "IMP-47422"},
		"env fields":        {document(importkind.XenonEnvironment, "xenon: 1\nvariables: x\n"), "IMP-47422"},
		"bundle syntax":     {document(importkind.XenonBundle, "item: [\n"), "IMP-47422"},
		"bundle fields":     {document(importkind.XenonBundle, "xenon: 1\nitem: x\n"), "IMP-47422"},
		"bundle kind":       {document(importkind.XenonBundle, "xenon: 1\nitem:\n  kind: environment\n"), "IMP-46422"},
		"bundle header":     {document(importkind.XenonBundle, "xenon: 1\nitem:\n  xenon: one\n"), "IMP-47422"},
		"bundle request":    {document(importkind.XenonBundle, "xenon: 1\nitem:\n  kind: request\n  headers: x\n"), "IMP-47422"},
		"bundle collection": {document(importkind.XenonBundle, "xenon: 1\nitem:\n  kind: collection\n  headers: x\n"), "IMP-47422"},
		"bundle children":   {document(importkind.XenonBundle, "xenon: 1\nitem:\n  kind: collection\n  children: x\n"), "IMP-47422"},
		"bundle child":      {document(importkind.XenonBundle, "xenon: 1\nitem:\n  kind: collection\n  children:\n    - kind: workspace\n"), "IMP-46422"},
		"zip no dir":        {entity.ImportInput{Kind: importkind.XenonZip, Probe: entity.ImportProbe{Name: "z", Entries: []string{exportManifestFile}}}, "IMP-48422"},
		"zip two dirs":      {entity.ImportInput{Kind: importkind.XenonZip, Probe: entity.ImportProbe{Entries: []string{"a/x", "b/y", "a/z"}}}, "IMP-48422"},
		"workspace no dir":  {entity.ImportInput{Kind: importkind.XenonWorkspaceZip, Probe: entity.ImportProbe{Entries: []string{"x"}}}, "IMP-48422"},
	}
	for name, tc := range cases {
		_, err := persistence.ReadXenonExport(ctx, tc.input)
		custom, isCustom := err.(errs.BaseError)
		if !isCustom || custom.Code() != tc.code {
			t.Errorf("%s: error = %v, want %s", name, err, tc.code)
		}
	}
	if _, err := persistence.ReadXenonExport(ctx, entity.ImportInput{Kind: importkind.XenonWorkspaceFolder, Root: filepath.Join(t.TempDir(), "missing")}); err == nil {
		t.Error("a missing workspace folder must fail")
	}
}

func TestReadXenonZipFailures(t *testing.T) {
	ctx := context.Background()
	persistence := NewXenonExportPersistence(NewWorkspaceIndexCache())
	zip := t.TempDir()
	input := entity.ImportInput{Kind: importkind.XenonZip, Root: zip, Probe: entity.ImportProbe{XenonRoot: "collection", Entries: []string{"shop/collection.yaml"}}}
	_, err := persistence.ReadXenonExport(ctx, input)
	expectErrorCode(t, err, "IMP-47422")
	writeTestFile(t, filepath.Join(zip, "shop", "collection.yaml"), exportCollection)
	writeTestFile(t, filepath.Join(zip, "shop", "b.yaml"), "name: [\n")
	_, err = persistence.ReadXenonExport(ctx, input)
	expectErrorCode(t, err, "IMP-47422")
	writeTestFile(t, filepath.Join(zip, "shop", "b.yaml"), exportSecond)
	writeTestFile(t, filepath.Join(zip, "shop", "products", "folder.yaml"), "xenon: 1\nkind: request\n")
	_, err = persistence.ReadXenonExport(ctx, input)
	expectErrorCode(t, err, "IMP-47422")
	writeTestFile(t, filepath.Join(zip, "shop", "products", "folder.yaml"), exportFolder)
	if err := os.Mkdir(filepath.Join(zip, exportSecretsFile), 0o755); err != nil {
		t.Fatal(err)
	}
	_, err = persistence.ReadXenonExport(ctx, input)
	expectErrorCode(t, err, "IMP-49500")
	workspace := entity.ImportInput{Kind: importkind.XenonWorkspaceZip, Root: zip, Probe: entity.ImportProbe{Entries: []string{"shop/collection.yaml"}}}
	_, err = persistence.ReadXenonExport(ctx, workspace)
	expectErrorCode(t, err, "IMP-49500")
}

func TestCopyFolderLeavesLocalOnlyFilesBehind(t *testing.T) {
	ctx := context.Background()
	source := writeExportWorkspace(t)
	persistence := NewWorkspacePersistence(NewWorkspaceIndexCache())
	target := filepath.Join(t.TempDir(), "copy")
	if err := os.Mkdir(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := persistence.CopyFolder(ctx, source, target); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"workspace.yaml", ".gitignore", "collections/shop/products/a.yaml", "files/.git", "files/payload.bin", "collections/shop/products/.hidden.yaml"} {
		copied, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(name)))
		original, _ := os.ReadFile(filepath.Join(source, filepath.FromSlash(name)))
		if err != nil || string(copied) != string(original) {
			t.Errorf("%s: %q, %v", name, copied, err)
		}
	}
	for _, name := range []string{".xenon", ".git", ".xenon-import-1", "collections/shop/.a.yaml.xenon-tmp"} {
		if _, err := os.Stat(filepath.Join(target, filepath.FromSlash(name))); !os.IsNotExist(err) {
			t.Errorf("%s was copied", name)
		}
	}
	expectOnly(t, filepath.Dir(target), "copy")
	fresh := filepath.Join(t.TempDir(), "new", "copy")
	if err := persistence.CopyFolder(ctx, source, fresh); err != nil {
		t.Fatal(err)
	}
	expectOnly(t, filepath.Dir(fresh), "copy")
}

func TestCopyFolderFailures(t *testing.T) {
	ctx := context.Background()
	persistence := NewWorkspacePersistence(NewWorkspaceIndexCache())
	parent := t.TempDir()
	err := persistence.CopyFolder(ctx, filepath.Join(parent, "missing"), filepath.Join(parent, "copy"))
	expectErrorCode(t, err, "WKS-33500")
	expectOnly(t, parent)
	source := writeExportWorkspace(t)
	writeTestFile(t, filepath.Join(parent, "taken", "x.txt"), "x")
	err = persistence.CopyFolder(ctx, source, filepath.Join(parent, "taken"))
	expectErrorCode(t, err, "WKS-34500")
	expectOnly(t, parent, "taken")
	writeTestFile(t, filepath.Join(parent, "file"), "x")
	err = persistence.CopyFolder(ctx, source, filepath.Join(parent, "file", "copy"))
	expectErrorCode(t, err, "WKS-33500")
}

func TestRemoveFolderRemovesTheCopiedWorkspace(t *testing.T) {
	ctx := context.Background()
	persistence := NewWorkspacePersistence(NewWorkspaceIndexCache())
	parent := t.TempDir()
	writeTestFile(t, filepath.Join(parent, "copy", "collections", "a.yaml"), "x")
	if err := persistence.RemoveFolder(ctx, filepath.Join(parent, "copy")); err != nil {
		t.Fatal(err)
	}
	expectOnly(t, parent)
	if err := persistence.RemoveFolder(ctx, filepath.Join(parent, "copy")); err != nil {
		t.Fatal(err)
	}
}

func TestReadImportSourceProbesXenonExports(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, "r.yaml"), exportRequest)
	writeTestFile(t, filepath.Join(dir, "b.YML"), "xenon: 1\nkind: bundle\nroot: folder\n")
	writeTestFile(t, filepath.Join(dir, "bad.yaml"), "xenon: [\n")
	writeTestFile(t, filepath.Join(dir, "r.json"), `{"xenon": 1, "kind": "request"}`)
	writeTestFile(t, filepath.Join(dir, "zip", exportManifestFile), "xenon: 1\nkind: export\nroot: workspace\n")
	writeTestFile(t, filepath.Join(dir, "zip", workspaceFile), "xenon: 1\nkind: workspace\n")
	writeTestFile(t, filepath.Join(dir, "ws", workspaceFile), "xenon: 2\nkind: workspace\n")
	writeTestFile(t, filepath.Join(dir, "plain", "nested", workspaceFile), "xenon: 1\nkind: workspace\n")
	persistence := NewImportSourcePersistence(nil)
	cases := map[string]entity.ImportProbe{
		"r.yaml":   {XenonFormat: 1, XenonKind: "request"},
		"b.YML":    {XenonFormat: 1, XenonKind: "bundle", XenonRoot: "folder"},
		"bad.yaml": {},
		"r.json":   {},
		"zip":      {XenonFormat: 1, XenonKind: "export", XenonRoot: "workspace"},
		"ws":       {XenonFormat: 2, XenonKind: "workspace"},
		"plain":    {},
	}
	for name, want := range cases {
		inputs, _, err := persistence.Read(context.Background(), entity.ImportRequest{Paths: []string{filepath.Join(dir, name)}})
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		probe := inputs[0].Probe
		if probe.XenonFormat != want.XenonFormat || probe.XenonKind != want.XenonKind || probe.XenonRoot != want.XenonRoot {
			t.Errorf("%s: probe = %+v", name, probe)
		}
	}
}
