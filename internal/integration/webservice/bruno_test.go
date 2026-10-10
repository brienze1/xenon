package webservice

import (
	"context"
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/envcolor"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
	"github.com/brienze1/xenon/pkg/errs"
)

func read(t *testing.T, kind enums.ImportKind, files map[string]string) *entity.BrunoCollection {
	t.Helper()
	input := entity.ImportInput{Kind: kind, Probe: entity.ImportProbe{Name: "folder-name"}}
	for path, content := range files {
		input.Documents = append(input.Documents, entity.ImportDocument{Path: path, Data: []byte(content)})
	}
	collection, err := NewBrunoWebservice().ReadCollection(context.Background(), input)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return collection
}

func names(nodes []entity.ImportNode) []string {
	result := make([]string, 0, len(nodes))
	for _, node := range nodes {
		result = append(result, node.Name())
	}
	return result
}

func TestReadCollectionOrdersBySequenceThenFileName(t *testing.T) {
	collection := read(t, importkind.BrunoBru, map[string]string{
		"b.bru":            "meta {\n  seq: 2\n}\nget {\n  url: https://b\n}",
		"a.bru":            "meta {\n  name: A\n  seq: 2\n}\npost {\n  url: https://a\n}",
		"first.bru":        "meta {\n  name: First\n  seq: 1\n}\nhttp {\n  method: purge\n}",
		"Loose/x.bru":      "meta {\n  name: X\n}\nget {\n  url: https://x\n}",
		"Loose/y.bru":      "meta {\n  name: Y\n  seq: 1\n}\ngrpc {\n  url: x\n}",
		"Typed/f.bru":      "meta {\n  name: F\n  seq: 1\n}\nws {\n  url: x\n}",
		"Typed/folder.bru": "meta {\n  seq: 0\n}",
		"notes.txt":        "ignored",
		"other.json":       "{}",
	})
	root := collection.Root
	if root.Name() != "folder-name" || !reflect.DeepEqual(names(root.Children), []string{"Typed", "First", "A", "b", "Loose"}) {
		t.Fatalf("children = %v", names(root.Children))
	}
	if root.Children[1].Request.Method != "PURGE" || root.Children[2].Request.Method != "POST" || root.Children[3].Request.Method != "GET" {
		t.Errorf("methods = %+v", root.Children[1].Request.Method)
	}
	if len(root.Children[0].Children) != 0 || !reflect.DeepEqual(names(root.Children[4].Children), []string{"X"}) {
		t.Errorf("folders = %v %v", names(root.Children[0].Children), names(root.Children[4].Children))
	}
	if root.Children[0].Container.Bruno["seq"] != 0 || root.Children[4].Container.Bruno != nil {
		t.Errorf("folder passthrough = %v %v", root.Children[0].Container.Bruno, root.Children[4].Container.Bruno)
	}
	warnings := collection.Warnings
	if len(warnings) != 2 || warnings[0].Code != warningcode.UnsupportedRequestType || warnings[0].Path != "folder-name/Typed" || warnings[0].ItemID != root.Children[0].ID() {
		t.Fatalf("warnings = %+v", warnings)
	}
	if warnings[0].Message != "F is a ws request, which Xenon does not support; it was skipped" || warnings[1].Message != "Y is a grpc request, which Xenon does not support; it was skipped" {
		t.Errorf("messages = %q %q", warnings[0].Message, warnings[1].Message)
	}
}

func TestReadCollectionMapsBruRequestDetails(t *testing.T) {
	collection := read(t, importkind.BrunoBru, map[string]string{
		"bruno.json": `{"name": "Shop", "version": "1"}`,
		"Sparql.bru": "meta {\n  name: Sparql\n}\npost {\n  url: u\n  body: sparql\n  auth: apikey\n}\nauth:apikey {\n  key: k\n  value: v\n  placement: queryparams\n}\nbody:sparql {\n  SELECT 1\n}\nsettings {\n  encodeUrl: false\n  followRedirects: true\n  timeout: abc\n  maxRedirects: 3\n  proxy: true\n  mode: fast\n}",
		"Files.bru":  "meta {\n  name: Files\n}\nput {\n  url: u\n  body: file\n  auth: wsse\n}\nauth:wsse {\n  username: a\n}\nbody:file {\n  ~file: @file(/a.bin) @contentType(text/plain)\n  file: @file(/b.bin)\n}",
		"Empty.bru":  "meta {\n  name: Empty\n}\nput {\n  url: u\n  body: file\n}\nbody:file {\n}",
	})
	if collection.Root.Name() != "Shop" || collection.Root.Container.Bruno["version"] != "1" || collection.Root.Container.Bruno["name"] != nil {
		t.Fatalf("collection = %+v", collection.Root.Container)
	}
	requests := map[string]entity.Request{}
	for _, node := range collection.Root.Children {
		requests[node.Name()] = *node.Request
	}
	sparql := requests["Sparql"]
	if sparql.Body.Mode != bodymode.Raw || sparql.Body.Raw.Content != "SELECT 1" || sparql.Bruno["bodyMode"] != "sparql" {
		t.Errorf("sparql body = %+v %v", sparql.Body, sparql.Bruno)
	}
	if sparql.Auth.Type != authtype.ApiKey || !reflect.DeepEqual(sparql.Auth.Attributes, []entity.AuthAttribute{{Key: "key", Value: "k"}, {Key: "value", Value: "v"}, {Key: "in", Value: "query"}}) {
		t.Errorf("apikey = %+v", sparql.Auth)
	}
	if sparql.Settings.EncodeUrl || !sparql.Settings.FollowRedirects || sparql.Settings.TimeoutMs != 0 {
		t.Errorf("settings = %+v", sparql.Settings)
	}
	if !reflect.DeepEqual(sparql.Bruno["settings"], map[string]any{"maxRedirects": 3, "proxy": true, "mode": "fast"}) {
		t.Errorf("settings passthrough = %v", sparql.Bruno["settings"])
	}
	files := requests["Files"]
	if files.Body.Binary.Src != "/b.bin" || !files.Body.Binary.NeedsRepick() || files.Bruno["file"] != nil {
		t.Errorf("selected file = %+v %v", files.Body.Binary, files.Bruno)
	}
	if files.Auth.Type != authtype.Unsupported || files.Auth.Unsupported.Type != "wsse" || !reflect.DeepEqual(files.Auth.Unsupported.Data, map[string]any{"username": "a"}) {
		t.Errorf("unsupported = %+v", files.Auth)
	}
	if empty := requests["Empty"]; empty.Body.Binary.Src != "" || empty.Body.Mode != bodymode.Binary {
		t.Errorf("empty file body = %+v", empty.Body)
	}
	if missing := countWarnings(collection.Warnings, warningcode.MissingFile); missing != 2 {
		t.Errorf("missing file warnings = %d %+v", missing, collection.Warnings)
	}
}

func TestReadCollectionMapsEnvironmentColours(t *testing.T) {
	collection := read(t, importkind.BrunoBru, map[string]string{
		"environments/a.bru":    "color: #0ff",
		"environments/b.bru":    "color: #ff00ff",
		"environments/c.bru":    "color: #ffee00",
		"environments/d.bru":    "color: #ff0000",
		"environments/e.bru":    "color: purple",
		"environments/f.bru":    "color: #12345",
		"environments/prod.bru": "vars {\n  a: 1\n}",
	})
	want := []enums.EnvColor{envcolor.Cyan, envcolor.Magenta, envcolor.Yellow, envcolor.Red, envcolor.None, envcolor.None, envcolor.Red}
	for index, environment := range collection.Environments {
		if environment.Environment.Color != want[index] {
			t.Errorf("%s: color = %s, want %s", environment.Environment.Name, environment.Environment.Color, want[index])
		}
	}
	if collection.Environments[0].Environment.Bruno["color"] != "#0ff" || collection.Environments[4].Environment.Bruno != nil {
		t.Errorf("passthrough = %v %v", collection.Environments[0].Environment.Bruno, collection.Environments[4].Environment.Bruno)
	}
	if !collection.Environments[6].Environment.ConfirmBeforeSend {
		t.Error("prod keeps its production defaults")
	}
}

func TestReadCollectionMapsYamlBodiesAndScripts(t *testing.T) {
	collection := read(t, importkind.BrunoYaml, map[string]string{
		"opencollection.yaml":  "opencollection: \"1.0.0\"\ninfo:\n  name: Y\nrequest:\n  variables:\n    - name: v\n      value: \"1\"\n      enabled: false\n    - name: after\n      value: res.status\n      scope: post-response",
		"Form.yml":             "info:\n  name: Form\n  seq: 1\nhttp:\n  method: POST\n  url: u\n  body:\n    type: form-urlencoded\n    data:\n      - name: a\n        value: \"1\"\n      - name: b\n        value: \"2\"\n        disabled: true",
		"File.yml":             "info:\n  name: File\n  seq: 2\nhttp:\n  method: PUT\n  url: u\n  body:\n    type: file\n    data:\n      - filePath: /a\n        contentType: text/plain\n        selected: true",
		"Multi.yml":            "info:\n  name: Multi\n  seq: 3\nhttp:\n  method: POST\n  url: u\n  body:\n    type: multipart-form\n    data:\n      - name: one\n        type: file\n        value: /one.pdf\n      - name: t\n        type: text\n        value: x",
		"Script.yml":           "info:\n  name: Script\n  seq: 4\nhttp:\n  url: u\nruntime:\n  scripts:\n    - type: before-request\n      code: a();\n    - type: after-response\n      code: >\n        b();\n  assertions:\n    - expression: res.status\n      operator: eq\n      value: \"1\"\n      enabled: false\ndocs: |\n  Doc\n",
		"Plain/folder.yaml":    "request:\n  auth:\n    type: none",
		"environments/dev.yml": "variables:\n  - name: k\n    value: v",
		"readme.md":            "x",
	})
	root := collection.Root
	if root.Name() != "Y" || root.Container.Bruno["opencollection"] != "1.0.0" || len(root.Container.Variables) != 1 || root.Container.Variables[0].Enabled {
		t.Fatalf("collection = %+v", root.Container)
	}
	if !reflect.DeepEqual(names(root.Children), []string{"Form", "File", "Multi", "Script", "Plain"}) {
		t.Fatalf("children = %v", names(root.Children))
	}
	if form := root.Children[0].Request.Body; form.Mode != bodymode.UrlEncoded || len(form.UrlEncoded) != 2 || form.UrlEncoded[1].Enabled {
		t.Errorf("form = %+v", form)
	}
	file := root.Children[1].Request
	if file.Body.Binary.Src != "/a" || !reflect.DeepEqual(file.Bruno["file"], map[string]any{"contentType": "text/plain"}) {
		t.Errorf("file = %+v %v", file.Body, file.Bruno)
	}
	if multi := root.Children[2].Request.Body.FormData; len(multi) != 2 || multi[0].Src.Src != "/one.pdf" || multi[1].Value != "x" {
		t.Errorf("multi = %+v", multi)
	}
	if script := root.Children[3].Request; script.Method != "GET" || script.Description != "Doc" {
		t.Errorf("script request = %+v", script)
	}
	if plain := root.Children[4].Container; plain.Name != "Plain" || plain.Auth.Type != authtype.None {
		t.Errorf("plain folder = %+v", plain)
	}
	if len(collection.Scripts) != 2 {
		t.Fatalf("scripts = %+v", collection.Scripts)
	}
	scripts := collection.Scripts[1]
	if scripts.PreRequest != (entity.BrunoScript{Text: "a();", Line: 9}) || scripts.PostResponse != (entity.BrunoScript{Text: "b();", Line: 12}) {
		t.Errorf("scripts = %+v", scripts)
	}
	if len(scripts.Assertions) != 1 || scripts.Assertions[0].Enabled || scripts.Assertions[0].Line != 14 {
		t.Errorf("assertions = %+v", scripts.Assertions)
	}
	if collection.Scripts[0].ItemID != root.ID() || len(collection.Scripts[0].PostResponseVars) != 1 {
		t.Errorf("collection scripts = %+v", collection.Scripts[0])
	}
	if dev := collection.Environments[0].Environment; dev.Name != "dev" || len(dev.Variables) != 1 {
		t.Errorf("environment = %+v", dev)
	}
}

func TestReadCollectionRefusesBrokenDocuments(t *testing.T) {
	cases := map[string]struct {
		kind  enums.ImportKind
		files map[string]string
		code  string
	}{
		"bruno.json":     {importkind.BrunoBru, map[string]string{"bruno.json": "{"}, "IMP-38400"},
		"bru file":       {importkind.BrunoBru, map[string]string{"a.bru": "meta {"}, "IMP-36400"},
		"yaml request":   {importkind.BrunoYaml, map[string]string{"a.yml": "info: [1"}, "IMP-37400"},
		"yaml folder":    {importkind.BrunoYaml, map[string]string{"x/folder.yml": "info: [1"}, "IMP-37400"},
		"yaml env":       {importkind.BrunoYaml, map[string]string{"environments/a.yml": "name: [1"}, "IMP-37400"},
		"opencollection": {importkind.BrunoYaml, map[string]string{"opencollection.yml": "info: [1"}, "IMP-37400"},
	}
	for name, tc := range cases {
		input := entity.ImportInput{Kind: tc.kind}
		for path, content := range tc.files {
			input.Documents = append(input.Documents, entity.ImportDocument{Path: path, Data: []byte(content)})
		}
		_, err := NewBrunoWebservice().ReadCollection(context.Background(), input)
		if custom, isCustom := err.(errs.BaseError); !isCustom || custom.Code() != tc.code {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

func countWarnings(warnings []entity.ImportWarning, code enums.WarningCode) int {
	count := 0
	for _, warning := range warnings {
		if warning.Code == code {
			count++
		}
	}
	return count
}
