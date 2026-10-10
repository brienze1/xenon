package usecase

import (
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
		"text":        {entity.ImportProbe{IsPasted: true, Name: "pasted text", Text: "curling"}, "IMP-00415", "pasted text is not a Postman, Bruno or cURL export"},
		"v1 shape":    {entity.ImportProbe{Name: "v1.json", IsJson: true, HasV1Shape: true}, "IMP-01400", "Postman v1 collections are not supported; export the collection as v2.1"},
		"v1 schema":   {entity.ImportProbe{Name: "v1.json", IsJson: true, Schema: "https://schema.getpostman.com/json/collection/v1.0.0/collection.json"}, "IMP-01400", "Postman v1 collections are not supported; export the collection as v2.1"},
		"v3 file":     {entity.ImportProbe{Name: "Get.Request.yaml"}, "IMP-00415", "Postman v3 YAML is not supported yet; export the collection as v2.1 JSON"},
		"v3 folder":   {entity.ImportProbe{Name: "v3", IsFolder: true, Entries: []string{".resources/definition.yaml", "users/Get.request.yaml"}}, "IMP-00415", "Postman v3 YAML is not supported yet; export the collection as v2.1 JSON"},
		"malformed":   {entity.ImportProbe{Name: "a.json", IsMalformed: true}, "IMP-01400", "a.json is not valid JSON"},
		"random json": {entity.ImportProbe{Name: "r.json", IsJson: true, Scope: "other"}, "IMP-00415", "r.json is not a Postman, Bruno or cURL export"},
		"empty dir":   {entity.ImportProbe{Name: "empty", IsFolder: true, Entries: []string{"sub/archive.json", "sub/bruno.json"}}, "IMP-00415", "empty is not a Postman, Bruno or cURL export"},
		"mixed":       {entity.ImportProbe{Name: "mixed", IsFolder: true, Entries: []string{"bruno.json", "opencollection.yml"}}, "IMP-00415", "mixed mixes Bruno .bru and YAML files"},
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
