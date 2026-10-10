package persistence

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
	"github.com/brienze1/xenon/pkg/errs"
)

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func errorCode(err error) string {
	if coded, ok := err.(errs.BaseError); ok {
		return coded.Code()
	}
	return ""
}

func skipWhenRoot(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("permission checks need a non-root unix user")
	}
}

func TestEnvironmentListReadsEveryYamlFileAndMarksBrokenOnes(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, environmentsDir)
	writeTestFile(t, filepath.Join(dir, "dev.yaml"), "xenon: 1\nkind: environment\nid: 01D\nname: dev\ncolor: green\nvariables:\n  - key: base_url\n    value: https://dev\n")
	writeTestFile(t, filepath.Join(dir, "PROD.YAML"), "id: 01P\nname: prod\nconfirmBeforeSend: true\n")
	writeTestFile(t, filepath.Join(dir, "broken.yaml"), "id: [unclosed\n")
	writeTestFile(t, filepath.Join(dir, "noid.yaml"), "kind: environment\nname: nameless\n")
	writeTestFile(t, filepath.Join(dir, "request.yaml"), "kind: request\nid: 01R\n")
	writeTestFile(t, filepath.Join(dir, "newer.yaml"), "xenon: 2\nid: 01N\n")
	writeTestFile(t, filepath.Join(dir, "notes.txt"), "id: 01T\n")
	writeTestFile(t, filepath.Join(dir, ".hidden.yaml"), "id: 01H\n")
	writeTestFile(t, filepath.Join(dir, "folder.yaml", "inner.yaml"), "id: 01I\n")
	environments, err := NewEnvironmentPersistence(NewWorkspaceIndexCache()).List(context.Background(), root)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	byFile := map[string]entity.Environment{}
	for _, environment := range environments {
		byFile[environment.FileName] = environment
	}
	if len(environments) != 6 {
		t.Fatalf("List() = %d environments: %v", len(environments), byFile)
	}
	dev := byFile["dev.yaml"]
	if dev.ID != "01D" || dev.Color != envcolor.Green || dev.IsMalformed() || len(dev.Variables) != 1 || dev.Variables[0].InitialValue != "https://dev" {
		t.Errorf("dev = %+v", dev)
	}
	if prod := byFile["PROD.YAML"]; prod.ID != "01P" || !prod.ConfirmBeforeSend || prod.IsMalformed() {
		t.Errorf("prod = %+v", prod)
	}
	reasons := map[string]string{
		"broken.yaml":  "broken.yaml cannot be read",
		"noid.yaml":    "noid.yaml has no id",
		"request.yaml": "request.yaml is a request file, expected a environment",
		"newer.yaml":   "newer.yaml was written by a newer Xenon (format 2)",
	}
	for file, reason := range reasons {
		environment := byFile[file]
		if !environment.IsMalformed() || !strings.HasPrefix(environment.Error, reason) || environment.ID != "error:"+file {
			t.Errorf("%s = %+v, want error %q", file, environment, reason)
		}
	}
}

func TestEnvironmentListWithoutAFolderIsEmpty(t *testing.T) {
	environments, err := NewEnvironmentPersistence(NewWorkspaceIndexCache()).List(context.Background(), t.TempDir())
	if err != nil || environments == nil || len(environments) != 0 {
		t.Fatalf("List() = %v, %v", environments, err)
	}
}

func TestEnvironmentListReportsAnUnreadableFolder(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, environmentsDir), "not a folder")
	persistence := NewEnvironmentPersistence(NewWorkspaceIndexCache())
	if _, err := persistence.List(context.Background(), root); errorCode(err) != "ENV-05500" {
		t.Fatalf("List() error = %v", err)
	}
	if _, err := persistence.Find(context.Background(), root, "01D"); errorCode(err) != "ENV-05500" {
		t.Fatalf("Find() error = %v", err)
	}
	if err := persistence.Delete(context.Background(), root, "01D"); errorCode(err) != "ENV-05500" {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestEnvironmentFind(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, environmentsDir, "dev.yaml"), "id: 01D\nname: dev\n")
	writeTestFile(t, filepath.Join(root, environmentsDir, "broken.yaml"), "id: [\n")
	persistence := NewEnvironmentPersistence(NewWorkspaceIndexCache())
	found, err := persistence.Find(context.Background(), root, "01D")
	if err != nil || found.Name != "dev" || found.FileName != "dev.yaml" {
		t.Fatalf("Find() = %+v, %v", found, err)
	}
	for _, id := range []string{"01X", "error:broken.yaml", ""} {
		if _, err := persistence.Find(context.Background(), root, id); errorCode(err) != "ENV-01404" || !errs.IsNotFoundError(err) {
			t.Errorf("Find(%q) error = %v", id, err)
		}
	}
}

func TestEnvironmentSaveCreatesPatchesAndRenames(t *testing.T) {
	root := t.TempDir()
	index := NewWorkspaceIndexCache()
	persistence := NewEnvironmentPersistence(index)
	ctx := context.Background()
	writeTestFile(t, filepath.Join(root, environmentsDir, "other.yaml"), "id: 01O\nname: other\n")
	environment := entity.NewEnvironment("01E", "staging")
	environment.FileName = "staging.yaml"
	environment.Variables = []entity.VariableRow{entity.NewVariableRow("base_url", "https://staging"), {Key: "api_key", InitialValue: "leaked", Secret: true, Enabled: true}}
	saved, err := persistence.Save(ctx, root, environment)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	path := filepath.Join(root, environmentsDir, "staging.yaml")
	content, _ := os.ReadFile(path)
	if strings.Contains(string(content), "leaked") || !strings.Contains(string(content), "secret: true") {
		t.Fatalf("staging.yaml =\n%s", content)
	}
	if !index.IsOwnWrite(ctx, path, content) {
		t.Error("the write is recorded for the watcher")
	}
	environment.Variables[1].InitialValue = ""
	if !reflect.DeepEqual(*saved, environment) {
		t.Fatalf("Save() =\n%+v\nwant\n%+v", *saved, environment)
	}
	writeTestFile(t, path, "# kept comment\ncustom: kept\n"+string(content))
	environment.Description = "patched"
	if _, err := persistence.Save(ctx, root, environment); err != nil {
		t.Fatalf("Save() in place error = %v", err)
	}
	patched, _ := os.ReadFile(path)
	if !strings.Contains(string(patched), "# kept comment") || !strings.Contains(string(patched), "custom: kept") || !strings.Contains(string(patched), "description: patched") {
		t.Fatalf("patched staging.yaml =\n%s", patched)
	}
	environment.Name = "qa"
	environment.FileName = "qa.yaml"
	renamed, err := persistence.Save(ctx, root, environment)
	if err != nil || renamed.FileName != "qa.yaml" || renamed.Name != "qa" {
		t.Fatalf("Save() rename = %+v, %v", renamed, err)
	}
	listed, _ := persistence.List(ctx, root)
	files := []string{}
	for _, item := range listed {
		files = append(files, item.FileName)
	}
	if !reflect.DeepEqual(files, []string{"other.yaml", "qa.yaml"}) {
		t.Fatalf("after rename the folder holds %v", files)
	}
}

func TestEnvironmentSaveOverwritesAnUnparseableFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, environmentsDir, "dev.yaml")
	writeTestFile(t, path, "id: [broken\n")
	environment := entity.NewEnvironment("01D", "dev")
	environment.FileName = "dev.yaml"
	saved, err := NewEnvironmentPersistence(NewWorkspaceIndexCache()).Save(context.Background(), root, environment)
	if err != nil || saved.IsMalformed() || saved.ID != "01D" {
		t.Fatalf("Save() = %+v, %v", saved, err)
	}
}

func TestEnvironmentSaveReportsFileErrors(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, environmentsDir), "not a folder")
	environment := entity.NewEnvironment("01D", "dev")
	environment.FileName = "dev.yaml"
	persistence := NewEnvironmentPersistence(NewWorkspaceIndexCache())
	if _, err := persistence.Save(context.Background(), root, environment); errorCode(err) != "ENV-06500" {
		t.Fatalf("Save() into a file error = %v", err)
	}
	skipWhenRoot(t)
	readOnly := t.TempDir()
	dir := filepath.Join(readOnly, environmentsDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if _, err := persistence.Save(context.Background(), readOnly, environment); errorCode(err) != "ENV-08500" {
		t.Fatalf("Save() into a read-only folder error = %v", err)
	}
	if err := os.Chmod(dir, 0o311); err != nil {
		t.Fatal(err)
	}
	if _, err := persistence.Save(context.Background(), readOnly, environment); errorCode(err) != "ENV-05500" {
		t.Fatalf("Save() with an unlistable folder error = %v", err)
	}
}

func TestEnvironmentDelete(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, environmentsDir)
	writeTestFile(t, filepath.Join(dir, "dev.yaml"), "id: 01D\nname: dev\n")
	writeTestFile(t, filepath.Join(dir, "hml.yaml"), "id: 01H\nname: hml\n")
	persistence := NewEnvironmentPersistence(NewWorkspaceIndexCache())
	if err := persistence.Delete(context.Background(), root, "01D"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "dev.yaml")); !os.IsNotExist(err) {
		t.Error("dev.yaml is removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "hml.yaml")); err != nil {
		t.Error("other environments stay")
	}
	if err := persistence.Delete(context.Background(), root, "01D"); errorCode(err) != "ENV-01404" {
		t.Fatalf("Delete() of a missing environment error = %v", err)
	}
	skipWhenRoot(t)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })
	if err := persistence.Delete(context.Background(), root, "01H"); errorCode(err) != "ENV-10500" {
		t.Fatalf("Delete() in a read-only folder error = %v", err)
	}
}
