package yamlnode

import (
	"testing"

	"go.yaml.in/yaml/v3"
)

func encode(t *testing.T, fields []Field, source string) string {
	t.Helper()
	document, err := Parse([]byte(source))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	Patch(Root(document), fields)
	out, err := Encode(document)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	return string(out)
}

func TestNewDocumentWritesFieldsInContractOrder(t *testing.T) {
	fields := []Field{
		Int("xenon", 1),
		String("kind", "workspace"),
		String("name", "Acme"),
		String("description", ""),
		StringList("order", []string{}),
		Reserved("variables", EmptyList()),
		Reserved("sync", EmptyMapping()),
		Bool("enabled", true),
	}
	want := "xenon: 1\nkind: workspace\nname: Acme\ndescription: \"\"\norder: []\nvariables: []\nsync: {}\nenabled: true\n"
	if got := encode(t, fields, ""); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestPatchPreservesUnknownFieldsAndComments(t *testing.T) {
	source := "# owned by xenon\nxenon: 1\nkind: workspace\ncustom: keep me # trailing\nname: Old # the name\nsync:\n  remote: origin\n"
	fields := []Field{
		Int("xenon", 1),
		String("kind", "workspace"),
		String("name", "New"),
		String("description", ""),
		Reserved("sync", EmptyMapping()),
	}
	want := "# owned by xenon\nxenon: 1\nkind: workspace\ncustom: keep me # trailing\nname: New # the name\ndescription: \"\"\nsync:\n  remote: origin\n"
	if got := encode(t, fields, source); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestPatchInsertsMissingFieldsBeforeTheFirstKnownOne(t *testing.T) {
	source := "custom: 1\nname: Acme\n"
	fields := []Field{String("kind", "workspace"), String("name", "Acme")}
	want := "kind: workspace\ncustom: 1\nname: Acme\n"
	if got := encode(t, fields, source); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestPatchMergesNestedObjectsKeepingOtherSubKeys(t *testing.T) {
	source := "auth:\n  type: apikey\n  apikey:\n    key: X-Key\n"
	fields := []Field{Object("auth", String("type", "bearer"), Object("bearer", String("token", "abc")))}
	want := "auth:\n  type: bearer\n  bearer:\n    token: abc\n  apikey:\n    key: X-Key\n"
	if got := encode(t, fields, source); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestPatchReplacesNonMappingValueWithObject(t *testing.T) {
	fields := []Field{Object("auth", String("type", "none"))}
	if got := encode(t, fields, "auth: broken\n"); got != "auth:\n  type: none\n" {
		t.Fatalf("got:\n%s", got)
	}
}

func TestEmptyObjectIsWrittenInFlowStyle(t *testing.T) {
	if got := encode(t, []Field{Object("postman")}, ""); got != "postman: {}\n" {
		t.Fatalf("got:\n%s", got)
	}
}

func TestStringsAreQuotedWhenAmbiguousAndMultilineUsesLiteralStyle(t *testing.T) {
	fields := []Field{
		String("flag", "true"),
		String("number", "42"),
		String("content", "{\n  \"amount\": 1000\n}\n"),
		StringList("order", []string{"a.yaml", "b"}),
	}
	want := "flag: \"true\"\nnumber: \"42\"\ncontent: |\n  {\n    \"amount\": 1000\n  }\norder:\n  - a.yaml\n  - b\n"
	if got := encode(t, fields, ""); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestListOfMappings(t *testing.T) {
	row := MappingNode(String("key", "Accept"), Bool("enabled", true))
	want := "headers:\n  - key: Accept\n    enabled: true\n"
	if got := encode(t, []Field{List("headers", []*yaml.Node{row})}, ""); got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestParseReadsBomAndCrlf(t *testing.T) {
	document, err := Parse([]byte("\xEF\xBB\xBFname: Acme\r\nkind: workspace\r\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	var decoded struct {
		Name string `yaml:"name"`
		Kind string `yaml:"kind"`
	}
	if err := document.Decode(&decoded); err != nil || decoded.Name != "Acme" || decoded.Kind != "workspace" {
		t.Fatalf("decoded = %+v, %v", decoded, err)
	}
}

func TestParseRejectsInvalidDocuments(t *testing.T) {
	for _, source := range []string{"- a\n- b\n", "key: [unclosed\n", "plain scalar\n"} {
		if _, err := Parse([]byte(source)); err == nil {
			t.Errorf("Parse(%q) must fail", source)
		}
	}
}
