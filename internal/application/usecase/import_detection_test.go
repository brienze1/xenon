package usecase

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/importkind"
	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestDetectImportFormatRecognisesEverySupportedSource(t *testing.T) {
	cases := map[string]struct {
		probe   entity.ImportProbe
		kind    enums.ImportKind
		version string
	}{
		"curl":             {entity.ImportProbe{IsPasted: true, Text: "  CURL -X POST https://x"}, importkind.Curl, ""},
		"v2.1":             {entity.ImportProbe{IsJson: true, Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"}, importkind.PostmanCollection, "2.1"},
		"v2.0":             {entity.ImportProbe{IsJson: true, Schema: "https://schema.getpostman.com/json/collection/v2.0.0/collection.json"}, importkind.PostmanCollection, "2.0"},
		"environment":      {entity.ImportProbe{IsJson: true, Scope: "environment"}, importkind.PostmanEnvironment, ""},
		"globals":          {entity.ImportProbe{IsJson: true, Scope: "globals"}, importkind.PostmanGlobals, ""},
		"dump":             {entity.ImportProbe{IsFolder: true, Entries: []string{"Archive.json", "collection/a.json"}}, importkind.PostmanDump, ""},
		"bruno":            {entity.ImportProbe{IsFolder: true, Entries: []string{"bruno.json", "a.bru"}}, importkind.BrunoBru, ""},
		"opencollection":   {entity.ImportProbe{IsFolder: true, Entries: []string{"opencollection.yml"}}, importkind.BrunoYaml, ""},
		"opencollection 2": {entity.ImportProbe{IsFolder: true, Entries: []string{"x/opencollection.yml", "opencollection.yaml"}}, importkind.BrunoYaml, ""},
		"xenon ws zip":     {entity.ImportProbe{IsFolder: true, XenonFormat: 1, XenonKind: "export", XenonRoot: "workspace"}, importkind.XenonWorkspaceZip, "1"},
		"xenon zip":        {entity.ImportProbe{IsFolder: true, XenonFormat: 1, XenonKind: "export", XenonRoot: "folder"}, importkind.XenonZip, "1"},
		"xenon bundle":     {entity.ImportProbe{XenonFormat: 1, XenonKind: "bundle", XenonRoot: "collection"}, importkind.XenonBundle, "1"},
		"xenon ws folder":  {entity.ImportProbe{IsFolder: true, XenonFormat: 1, XenonKind: "workspace"}, importkind.XenonWorkspaceFolder, "1"},
		"xenon request":    {entity.ImportProbe{XenonFormat: 1, XenonKind: "request"}, importkind.XenonRequest, "1"},
		"xenon env":        {entity.ImportProbe{XenonFormat: 1, XenonKind: "environment"}, importkind.XenonEnvironment, "1"},
		"v3 request":       {entity.ImportProbe{Name: "Get.Request.yaml"}, importkind.PostmanRequest, "3"},
		"v3 request kind":  {entity.ImportProbe{Name: "x.yaml", IsYaml: true, PostmanKind: "http-request"}, importkind.PostmanRequest, "3"},
		"v3 folder":        {entity.ImportProbe{Name: "v3", IsFolder: true, Entries: []string{".resources/definition.yaml", "users/Get.request.yaml"}}, importkind.PostmanCollection, "3"},
		"v3 definition":    {entity.ImportProbe{Name: "c", IsFolder: true, Entries: []string{".resources/definition.yaml"}}, importkind.PostmanCollection, "3"},
		"v3 json":          {entity.ImportProbe{Name: "c.json", IsJson: true, PostmanKind: "collection", HasItems: true}, importkind.PostmanCollection, "3"},
		"v3 yaml def":      {entity.ImportProbe{Name: "definition.yaml", IsYaml: true, PostmanKind: "collection"}, importkind.PostmanCollection, "3"},
		"v3 env yaml":      {entity.ImportProbe{Name: "dev.environment.yaml", IsYaml: true, HasValues: true}, importkind.PostmanEnvironment, "3"},
		"v3 env json":      {entity.ImportProbe{Name: "dev.json", IsJson: true, HasValues: true}, importkind.PostmanEnvironment, "3"},
		"v3 globals yaml":  {entity.ImportProbe{Name: "workspace.globals.yaml", IsYaml: true, HasValues: true}, importkind.PostmanGlobals, "3"},
		"v3 cloud globals": {entity.ImportProbe{Name: "g.json", IsJson: true, HasValues: true, HasWorkspaceField: true}, importkind.PostmanGlobals, "3"},
		"v3 local globals": {entity.ImportProbe{Name: "g.json", IsJson: true, HasValues: true, DocumentID: "postman/globals/workspace.globals.yaml"}, importkind.PostmanGlobals, "3"},
		"v3 workspace":     {entity.ImportProbe{Name: "repo", IsFolder: true, Entries: []string{"collections/A/x.request.yaml", "specs/a.yaml"}}, importkind.PostmanWorkspace, "3"},
		"v3 pulled":        {entity.ImportProbe{Name: "repo", IsFolder: true, Entries: []string{"postman/environments/dev.environment.yaml"}}, importkind.PostmanWorkspace, "3"},
		"v3 globals only":  {entity.ImportProbe{Name: "p", IsFolder: true, Entries: []string{"globals/workspace.globals.yaml"}}, importkind.PostmanWorkspace, "3"},
		"v3 collections":   {entity.ImportProbe{Name: "collections", IsFolder: true, Entries: []string{"A/x.request.yaml", "B/.resources/definition.yaml"}}, importkind.PostmanWorkspace, "3"},
		"specs only":       {entity.ImportProbe{Name: "s", IsFolder: true, Entries: []string{"postman/specs/a.yaml"}}, importkind.Folder, ""},
		"empty dir":        {entity.ImportProbe{Name: "empty", IsFolder: true, Entries: []string{"sub/archive.json", "sub/bruno.json"}}, importkind.Folder, ""},
	}
	for name, tc := range cases {
		kind, version, err := DetectImportFormat(tc.probe)
		if err != nil || kind != tc.kind || version != tc.version {
			t.Errorf("%s: got %q %q %v, want %q %q", name, kind, version, err, tc.kind, tc.version)
		}
	}
}

func TestDetectImportFormatRefusesWithAReason(t *testing.T) {
	cases := map[string]struct {
		probe   entity.ImportProbe
		code    string
		message string
	}{
		"text":        {entity.ImportProbe{IsPasted: true, Name: "pasted text", Text: "curling"}, "IMP-00415", "pasted text is not a Xenon, Postman, Bruno or cURL export"},
		"v1 shape":    {entity.ImportProbe{Name: "v1.json", IsJson: true, HasV1Shape: true}, "IMP-01400", "Postman v1 collections are not supported; export the collection as v2.1"},
		"v1 schema":   {entity.ImportProbe{Name: "v1.json", IsJson: true, Schema: "https://schema.getpostman.com/json/collection/v1.0.0/collection.json"}, "IMP-01400", "Postman v1 collections are not supported; export the collection as v2.1"},
		"other yaml":  {entity.ImportProbe{Name: "notes.yaml", IsYaml: true}, "IMP-00415", "notes.yaml is not a Xenon, Postman, Bruno or cURL export"},
		"malformed":   {entity.ImportProbe{Name: "a.json", IsMalformed: true}, "IMP-01400", "a.json is not valid JSON"},
		"random json": {entity.ImportProbe{Name: "r.json", IsJson: true, Scope: "other"}, "IMP-00415", "r.json is not a Xenon, Postman, Bruno or cURL export"},
		"mixed":       {entity.ImportProbe{Name: "mixed", IsFolder: true, Entries: []string{"bruno.json", "opencollection.yml"}}, "IMP-00415", "mixed mixes Bruno .bru and YAML files"},
		"newer":       {entity.ImportProbe{Name: "n.yaml", XenonFormat: 2, XenonKind: "request"}, "IMP-42422", "n.yaml was written by a newer Xenon (format 2); update Xenon to import it"},
		"ws file":     {entity.ImportProbe{Name: "workspace.yaml", XenonFormat: 1, XenonKind: "workspace"}, "IMP-00415", "workspace.yaml is not a Xenon, Postman, Bruno or cURL export"},
		"req folder":  {entity.ImportProbe{Name: "r", IsFolder: true, XenonFormat: 1, XenonKind: "request"}, "IMP-00415", "r is not a Xenon, Postman, Bruno or cURL export"},
		"env folder":  {entity.ImportProbe{Name: "e", IsFolder: true, XenonFormat: 1, XenonKind: "environment"}, "IMP-00415", "e is not a Xenon, Postman, Bruno or cURL export"},
		"no kind":     {entity.ImportProbe{Name: "k.yaml", XenonFormat: 1}, "IMP-00415", "k.yaml is not a Xenon, Postman, Bruno or cURL export"},
		"no format":   {entity.ImportProbe{Name: "f.yaml", XenonKind: "request"}, "IMP-00415", "f.yaml is not a Xenon, Postman, Bruno or cURL export"},
	}
	for name, tc := range cases {
		_, _, err := DetectImportFormat(tc.probe)
		custom, isCustom := err.(errs.BaseError)
		if !isCustom || custom.Code() != tc.code || custom.Description() != tc.message {
			t.Errorf("%s: got %v, want %s %q", name, err, tc.code, tc.message)
		}
	}
	if _, _, err := DetectImportFormat(entity.ImportProbe{Name: "x", IsPasted: false, Text: "curl https://x"}); err == nil {
		t.Error("only pasted text is checked for a cURL command")
	}
}

func TestRejectUnimportableFolderListsTheAcceptedLayouts(t *testing.T) {
	err := RejectUnimportableFolder("x")
	custom, isCustom := err.(errs.BaseError)
	want := "x has no file Xenon can import; pass a Postman v3 workspace or collection folder, a Postman data dump (archive.json), a Bruno collection (bruno.json or opencollection.yml), a Xenon workspace folder, or a folder of standalone export files"
	if !isCustom || custom.Code() != "IMP-55415" || custom.StatusCode() != http.StatusUnsupportedMediaType || custom.Description() != want {
		t.Errorf("got %v", err)
	}
}

func TestScanPmUsageReportsApisOutsideTheSubset(t *testing.T) {
	scripts := []entity.ScriptRef{
		{ItemID: "c", Path: "Acme", Phase: scriptphase.PostResponse, Text: "pm.test(\"ok\", () => pm.response.to.have.status(200));"},
		{ItemID: "r", Path: "Acme/Raw", Phase: scriptphase.PreRequest, Text: "// pm.visualizer.set\nconst jar = pm.cookies.jar();\nconsole.log(pm.info.requestName, \"pm.iterationData\");\n/* pm.globals\n pm.cookies.x */ pm.require('x');\nx.pm.ignored; pm.request.url.toString(); pm.response; pm.expect(1)"},
		{ItemID: "t", Path: "Acme/T", Phase: scriptphase.PreRequest, Text: "const a = `${pm.cookies.jar}`;\nconst b = 'it\\'s pm.info';\npm.execution.setNextRequest(null); pm"},
	}
	usages := ScanPmUsage(scripts)
	want := []entity.PmUsage{
		{API: "pm.cookies.jar", Occurrences: []entity.PmOccurrence{{ItemID: "r", Path: "Acme/Raw", Phase: scriptphase.PreRequest, Line: 2}}},
		{API: "pm.execution.setNextRequest", Occurrences: []entity.PmOccurrence{{ItemID: "t", Path: "Acme/T", Phase: scriptphase.PreRequest, Line: 3}}},
		{API: "pm.info.requestName", Occurrences: []entity.PmOccurrence{{ItemID: "r", Path: "Acme/Raw", Phase: scriptphase.PreRequest, Line: 3}}},
		{API: "pm.require", Occurrences: []entity.PmOccurrence{{ItemID: "r", Path: "Acme/Raw", Phase: scriptphase.PreRequest, Line: 5}}},
		{API: "pm.response.to", Occurrences: []entity.PmOccurrence{{ItemID: "c", Path: "Acme", Phase: scriptphase.PostResponse, Line: 1}}},
	}
	if !reflect.DeepEqual(usages, want) {
		t.Fatalf("ScanPmUsage() =\n%+v\nwant\n%+v", usages, want)
	}
	warnings := ListPmUsageWarnings(append(usages, entity.PmUsage{API: "pm.require", Occurrences: []entity.PmOccurrence{{ItemID: "r", Path: "Acme/Raw", Phase: scriptphase.PreRequest, Line: 9}}}))
	if len(warnings) != 5 {
		t.Fatalf("one warning per api and script, got %+v", warnings)
	}
	if warnings[0].Code != warningcode.PmApiOutsideMvp || warnings[0].Message != "The script uses pm.cookies.jar, which is outside Xenon's script subset" || warnings[0].Location != "preRequest:2" || warnings[0].ItemID != "r" || warnings[0].Path != "Acme/Raw" {
		t.Errorf("warning = %+v", warnings[0])
	}
	if warnings[3].Code != warningcode.PmRequire || warnings[3].Message != "The script uses pm.require, which only works inside Postman" {
		t.Errorf("pm.require warning = %+v", warnings[3])
	}
}

func TestScanPmUsageHandlesUnterminatedText(t *testing.T) {
	for _, text := range []string{"pm.cookies.jar // open", "/* pm.cookies.jar", "'pm.cookies.jar", "pm.", "pm.cookies.\n"} {
		usages := ScanPmUsage([]entity.ScriptRef{{Text: text}})
		if text == "pm.cookies.jar // open" || text == "pm.cookies.\n" {
			if len(usages) != 1 || usages[0].API != map[string]string{"pm.cookies.jar // open": "pm.cookies.jar", "pm.cookies.\n": "pm.cookies"}[text] {
				t.Errorf("%q: %+v", text, usages)
			}
			continue
		}
		if len(usages) != 0 {
			t.Errorf("%q: %+v", text, usages)
		}
	}
}
