package usecase

import (
	"reflect"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestTranslateBrunoScriptTranslatesEveryTableRow(t *testing.T) {
	cases := map[string]string{
		`bru.getEnvVar("a")`:             `pm.environment.get("a")`,
		`bru.setEnvVar("a", 1)`:          `pm.environment.set("a", 1)`,
		`bru.getVar("a")`:                `pm.variables.get("a")`,
		`bru.setVar("a", 1)`:             `pm.variables.set("a", 1)`,
		`bru.getRequestVar("a")`:         `pm.variables.get("a")`,
		`bru.getCollectionVar("a")`:      `pm.collectionVariables.get("a")`,
		`bru.getGlobalEnvVar("a")`:       `pm.globals.get("a")`,
		`bru.setGlobalEnvVar("a", 1)`:    `pm.globals.set("a", 1)`,
		`bru.getEnvName()`:               `pm.environment.name`,
		`bru.interpolate("{{x}}")`:       `pm.variables.replaceIn("{{x}}")`,
		`req.getUrl()`:                   `pm.request.url.toString()`,
		`req.url`:                        `pm.request.url.toString()`,
		`req.getMethod()`:                `pm.request.method`,
		`req.method`:                     `pm.request.method`,
		`req.getHeader("a")`:             `pm.request.headers.get("a")`,
		`req.setHeader("a", b(1, 2))`:    `pm.request.headers.upsert({ key: "a", value: b(1, 2) })`,
		`req.getHeaders()`:               `pm.request.headers.toObject()`,
		`req.headers["a"]`:               `pm.request.headers.toObject()["a"]`,
		`req.getBody()`:                  `JSON.parse(pm.request.body.raw)`,
		`req.body.a`:                     `JSON.parse(pm.request.body.raw).a`,
		`req.setBody({ a: [1, 2] })`:     `pm.request.body.update(JSON.stringify({ a: [1, 2] }))`,
		`res.getStatus()`:                `pm.response.code`,
		`res.status`:                     `pm.response.code`,
		`res.getBody().a`:                `pm.response.json().a`,
		`res.body.a`:                     `pm.response.json().a`,
		`res.getHeader("a")`:             `pm.response.headers.get("a")`,
		`res.getHeaders()`:               `pm.response.headers.toObject()`,
		`res.headers`:                    `pm.response.headers.toObject()`,
		`res.getResponseTime()`:          `pm.response.responseTime`,
		`res.responseTime`:               `pm.response.responseTime`,
		`test("a", () => {})`:            `pm.test("a", () => {})`,
		`expect(1).to.eql(1)`:            `pm.expect(1).to.eql(1)`,
		`req.setHeader("a", res.status)`: `pm.request.headers.upsert({ key: "a", value: pm.response.code })`,
		"`${res.status} ${req.url}`":     "`${pm.response.code} ${pm.request.url.toString()}`",
		"`{${ {a: res.status}.a }}`":     "`{${ {a: pm.response.code}.a }}`",
		"`a\\`b ${1}`":                   "`a\\`b ${1}`",
		`"res.status"`:                   `"res.status"`,
		`'a\'res.status'`:                `'a\'res.status'`,
		`pm.test(1); x.test(2)`:          `pm.test(1); x.test(2)`,
		`o.res.status`:                   `o.res.status`,
		`/* res.status */ res.status`:    `/* res.status */ pm.response.code`,
		`/* res.status`:                  `/* res.status`,
		`a = {} // res.status`:           `a = {} // res.status`,
		`{ x: res.status }`:              `{ x: pm.response.code }`,
		`resource.status`:                `resource.status`,
		`res.statusText`:                 `res.statusText`,
		`res . status`:                   `res . status`,
		`bru.getEnvName`:                 `bru.getEnvName`,
		`req.url()`:                      `req.url()`,
		`test`:                           `test`,
		`req.setHeader("a")`:             `req.setHeader("a")`,
		`req.setBody()`:                  `req.setBody()`,
		`req.setHeader("a", 1`:           `req.setHeader("a", 1`,
		`req.setHeader`:                  `req.setHeader`,
		`"unterminated`:                  `"unterminated`,
		"`unterminated":                  "`unterminated",
		"`${res.status":                  "`${pm.response.code",
	}
	for source, want := range cases {
		if got := translateBrunoCode(source); got != want {
			t.Errorf("%s: got %q, want %q", source, got, want)
		}
	}
}

func TestTranslateBrunoScriptCommentsOutAndReportsUntranslatedLines(t *testing.T) {
	script := entity.BrunoScript{Line: 10, Text: strings.Join([]string{
		"const a = res.status;",
		"  bru.sleep(1);",
		"const b = require('lodash');",
		"// bru.sleep(1);",
		"x.require('a'); res(\"x\");",
		"",
		"\treq.setUrl(\"u\");",
		"req.setHeader(\"a\",",
		"  res.status);",
		"/* bru.sleep(1)",
		"*/ const c = 1;",
	}, "\n")}
	got := TranslateBrunoScript(script)
	wantLines := []string{
		"const a = pm.response.code;",
		"  // XENON: untranslated bru.sleep(1);",
		"// XENON: untranslated const b = require('lodash');",
		"// bru.sleep(1);",
		"// XENON: untranslated x.require('a'); res(\"x\");",
		"",
		"\t// XENON: untranslated req.setUrl(\"u\");",
		"pm.request.headers.upsert({ key: \"a\", value: pm.response.code })",
		"/* bru.sleep(1)",
		"*/ const c = 1;",
	}
	wantLines = append(wantLines[:7], "pm.request.headers.upsert({ key: \"a\", value: pm.response.code })")
	if len(got.Lines) != 11 {
		t.Fatalf("lines = %d %q", len(got.Lines), got.Lines)
	}
	wantUntranslated := []entity.UntranslatedLine{
		{Line: 11, Source: "bru.sleep(1);"},
		{Line: 12, Source: "const b = require('lodash');"},
		{Line: 14, Source: `x.require('a'); res("x");`},
		{Line: 16, Source: `req.setUrl("u");`},
	}
	if !reflect.DeepEqual(got.Untranslated, wantUntranslated) {
		t.Errorf("untranslated = %+v, want %+v", got.Untranslated, wantUntranslated)
	}
	for index, want := range wantLines[:7] {
		if got.Lines[index] != want {
			t.Errorf("line %d = %q, want %q", index, got.Lines[index], want)
		}
	}
	if got.Lines[7] != "pm.request.headers.upsert({ key: \"a\", value: " || got.Lines[8] != "  pm.response.code });" {
		t.Errorf("multi-line call = %q %q", got.Lines[7], got.Lines[8])
	}
	if got.Lines[9] != "/* bru.sleep(1)" || got.Lines[10] != "*/ const c = 1;" {
		t.Errorf("block comment = %q %q", got.Lines[9], got.Lines[10])
	}
}

func TestTranslateBrunoScriptOfAnEmptyScriptHasNoLines(t *testing.T) {
	got := TranslateBrunoScript(entity.BrunoScript{Text: "  \n ", Line: 3})
	if len(got.Lines) != 0 || len(got.Untranslated) != 0 || got.Text() != "" {
		t.Errorf("got %+v", got)
	}
}

func TestConvertBrunoVarsBuildsVariableSetLines(t *testing.T) {
	pre := ConvertBrunoPreRequestVars([]entity.BrunoVariable{
		{Key: "host", Value: `{{base}} "q"`, Enabled: true},
		{Key: "off", Value: "x", Enabled: false},
	})
	if pre.Text() != "pm.variables.set(\"host\", \"{{base}} \\\"q\\\"\");\n// pm.variables.set(\"off\", \"x\");" || len(pre.Untranslated) != 0 {
		t.Errorf("pre = %q %+v", pre.Text(), pre.Untranslated)
	}
	post := ConvertBrunoPostResponseVars([]entity.BrunoVariable{
		{Key: "id", Value: " res.body.id ", Enabled: true, Line: 4},
		{Key: "raw", Value: `res("a")`, Enabled: true, Line: 5},
		{Key: "skip", Value: `res("a")`, Enabled: false, Line: 6},
	})
	want := strings.Join([]string{
		`pm.variables.set("id", pm.response.json().id);`,
		`// XENON: untranslated pm.variables.set("raw", res("a"));`,
		`// pm.variables.set("skip", res("a"));`,
	}, "\n")
	if post.Text() != want {
		t.Errorf("post = %q", post.Text())
	}
	if !reflect.DeepEqual(post.Untranslated, []entity.UntranslatedLine{{Line: 5, Source: `pm.variables.set("raw", res("a"));`}}) {
		t.Errorf("untranslated = %+v", post.Untranslated)
	}
}

func TestConvertBrunoAssertionsMapsEveryOperator(t *testing.T) {
	cases := map[string]struct{ operator, value, want string }{
		"eq number":      {"eq", "200", `.to.eql(200)`},
		"eq negative":    {"eq", "-1.5e3", `.to.eql(-1.5e3)`},
		"eq text":        {"eq", "ok", `.to.eql("ok")`},
		"eq quoted":      {"eq", `'ok'`, `.to.eql('ok')`},
		"eq bool":        {"eq", "true", `.to.eql(true)`},
		"eq null":        {"eq", "null", `.to.eql(null)`},
		"eq infinity":    {"eq", "Infinity", `.to.eql("Infinity")`},
		"neq":            {"neq", "1", `.to.not.eql(1)`},
		"gt":             {"gt", "1", `.to.be.above(1)`},
		"gte":            {"gte", "1", `.to.be.at.least(1)`},
		"lt":             {"lt", "1", `.to.be.below(1)`},
		"lte":            {"lte", "1", `.to.be.at.most(1)`},
		"in list":        {"in", "1, a", `.to.be.oneOf([1, "a"])`},
		"in array":       {"in", "[1,2]", `.to.be.oneOf([1,2])`},
		"notIn":          {"notIn", "1", `.to.not.be.oneOf([1])`},
		"contains":       {"contains", "a", `.to.include("a")`},
		"notContains":    {"notContains", "a", `.to.not.include("a")`},
		"length":         {"length", "2", `.to.have.lengthOf(2)`},
		"matches":        {"matches", "^a/b", `.to.match(/^a\/b/)`},
		"matches quoted": {"matches", `"^a"`, `.to.match(/^a/)`},
		"notMatches":     {"notMatches", "x", `.to.not.match(/x/)`},
		"startsWith":     {"startsWith", "a.b", `.to.match(/^a\.b/)`},
		"endsWith":       {"endsWith", `"a/b"`, `.to.match(/a\/b$/)`},
		"between":        {"between", "1, 2", `.to.be.within(1, 2)`},
		"isEmpty":        {"isEmpty", "", `.to.be.empty`},
		"isNotEmpty":     {"isNotEmpty", "", `.to.not.be.empty`},
		"isNull":         {"isNull", "", `.to.be.null`},
		"isUndefined":    {"isUndefined", "", `.to.be.undefined`},
		"isDefined":      {"isDefined", "", `.to.not.be.undefined`},
		"isTruthy":       {"isTruthy", "", `.to.be.ok`},
		"isFalsy":        {"isFalsy", "", `.to.not.be.ok`},
		"isNumber":       {"isNumber", "", `.to.be.a("number")`},
		"isString":       {"isString", "", `.to.be.a("string")`},
		"isBoolean":      {"isBoolean", "", `.to.be.a("boolean")`},
		"isArray":        {"isArray", "", `.to.be.a("array")`},
		"isJson":         {"isJson", "", `.to.be.an("object")`},
	}
	for name, tc := range cases {
		got := ConvertBrunoAssertions([]entity.BrunoAssertion{{Expression: "res.status", Operator: tc.operator, Value: tc.value, Enabled: true, Line: 7}})
		title := strings.TrimSpace("res.status " + strings.TrimSpace(tc.operator+" "+tc.value))
		want := "pm.test(" + jsString(title) + ", () => pm.expect(pm.response.code)" + tc.want + ");"
		if got.Text() != want || len(got.Untranslated) != 0 {
			t.Errorf("%s: got %q %+v, want %q", name, got.Text(), got.Untranslated, want)
		}
	}
}

func TestConvertBrunoAssertionsFlagsWhatCannotBeConverted(t *testing.T) {
	assertions := []entity.BrunoAssertion{
		{Expression: "res.size", Operator: "isHuge", Enabled: true, Line: 1},
		{Expression: `res("a")`, Operator: "eq", Value: "1", Enabled: true, Line: 2},
		{Expression: "res.status", Operator: "between", Value: "1", Enabled: true, Line: 3},
		{Expression: "res.status", Operator: "eq", Value: "", Enabled: true, Line: 4},
		{Expression: "res.status", Operator: "in", Value: "", Enabled: true, Line: 5},
		{Expression: "res.status", Operator: "matches", Value: "", Enabled: true, Line: 6},
		{Expression: "res.status", Operator: "isHuge", Enabled: false, Line: 7},
		{Expression: "res.status", Operator: "eq", Value: "1", Enabled: false, Line: 8},
	}
	got := ConvertBrunoAssertions(assertions)
	want := []string{
		"// XENON: untranslated res.size: isHuge",
		`// XENON: untranslated res("a"): eq 1`,
		"// XENON: untranslated res.status: between 1",
		"// XENON: untranslated res.status: eq",
		"// XENON: untranslated res.status: in",
		"// XENON: untranslated res.status: matches",
		"// res.status: isHuge",
		`// pm.test("res.status eq 1", () => pm.expect(pm.response.code).to.eql(1));`,
	}
	if !reflect.DeepEqual(got.Lines, want) {
		t.Errorf("lines = %q", got.Lines)
	}
	if len(got.Untranslated) != 6 || got.Untranslated[1] != (entity.UntranslatedLine{Line: 2, Source: `res("a"): eq 1`}) {
		t.Errorf("untranslated = %+v", got.Untranslated)
	}
}

func TestDeriveCurlRequestNameUsesTheMethodAndPath(t *testing.T) {
	cases := map[string]string{
		"https://api.test/users?x=1#f": "POST /users",
		"https://api.test/a/b/":        "POST /a/b",
		"https://api.test":             "POST api.test",
		"https://api.test/":            "POST api.test",
		"{{base}}/users/{{id}}":        "POST /users/{{id}}",
		"api.test/x#frag/y":            "POST /x",
	}
	for address, want := range cases {
		request := entity.NewRequest("", "")
		request.Method = "post"
		request.URL = address
		if got := DeriveCurlRequestName(request); got != want {
			t.Errorf("%s: got %q, want %q", address, got, want)
		}
	}
}

func TestIsCurlCommandAcceptsCurlAndCurlExe(t *testing.T) {
	cases := map[string]bool{
		"curl https://x":      true,
		"  CURL.EXE 'x'":      true,
		"curlx https://x":     false,
		"curl":                false,
		"wget https://x/curl": false,
	}
	for text, want := range cases {
		if got := IsCurlCommand(text); got != want {
			t.Errorf("%q: got %v", text, got)
		}
	}
}

func TestDetectImportFormatRefusesMixedBrunoFormats(t *testing.T) {
	cases := map[string]struct {
		entries []string
		mixed   bool
	}{
		"bru with yml request":       {[]string{"bruno.json", "a/B.yml"}, true},
		"bru with yaml request":      {[]string{"bruno.json", "B.YAML"}, true},
		"bru with hidden yml":        {[]string{"bruno.json", ".github/ci.yml", "a.bru"}, false},
		"yaml with bru request":      {[]string{"opencollection.yml", "a/A.BRU"}, true},
		"yaml with hidden bru":       {[]string{"opencollection.yml", "x/.old/A.bru"}, false},
		"yaml collection":            {[]string{"opencollection.yml", "a/B.yml"}, false},
		"bru collection with readme": {[]string{"bruno.json", "README.md", "a.bru"}, false},
	}
	for name, tc := range cases {
		_, _, err := DetectImportFormat(entity.ImportProbe{Name: "shop", IsFolder: true, Entries: tc.entries})
		if (err != nil) != tc.mixed {
			t.Errorf("%s: err = %v, want mixed %v", name, err, tc.mixed)
		}
		if custom, isCustom := err.(errs.BaseError); err != nil && (!isCustom || custom.Code() != "IMP-00415" || custom.Description() != "shop mixes Bruno .bru and YAML files") {
			t.Errorf("%s: error = %v", name, err)
		}
	}
}
