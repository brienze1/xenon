package bruno

import (
	"reflect"
	"strings"
	"testing"

	"github.com/brienze1/xenon/internal/integration/webservice/bruno/dto"
	"github.com/brienze1/xenon/pkg/errs"
)

func parse(t *testing.T, lines ...string) *dto.BruFile {
	t.Helper()
	file, err := parseBru("a.bru", []byte(strings.Join(lines, "\n")))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return file
}

func TestParseBruReadsDictionaryBlocks(t *testing.T) {
	file := parse(t,
		"meta {",
		"  name: List: all",
		"  seq: 2",
		"",
		"}",
		"headers {",
		"  Accept: application/json",
		"  ~X-Trace: 1  ",
		"  Empty:",
		`  "quoted: key": v`,
		`  ~"off": w`,
		"}",
	)
	if len(file.Blocks) != 2 || file.Blocks[0].Name != "meta" || file.Blocks[0].Line != 1 || file.Blocks[1].Line != 6 {
		t.Fatalf("blocks = %+v", file.Blocks)
	}
	want := []dto.BruEntry{
		{Key: "Accept", Value: "application/json", Line: 7, Enabled: true},
		{Key: "X-Trace", Value: "1", Line: 8, Enabled: false},
		{Key: "Empty", Value: "", Line: 9, Enabled: true},
		{Key: "quoted: key", Value: "v", Line: 10, Enabled: true},
		{Key: "off", Value: "w", Line: 11, Enabled: false},
	}
	if !reflect.DeepEqual(file.Entries("headers"), want) {
		t.Errorf("entries = %+v", file.Entries("headers"))
	}
	if file.Value("meta", "name") != "List: all" || file.Value("meta", "seq") != "2" || file.Value("headers", "X-Trace") != "" || file.Value("nope", "x") != "" {
		t.Errorf("values = %q %q", file.Value("meta", "name"), file.Value("meta", "seq"))
	}
}

func TestParseBruReadsTextBlocksWithTheirFirstLine(t *testing.T) {
	file := parse(t,
		"docs {",
		"  # Title",
		"",
		"    indented",
		"no indent",
		"  }",
		"}",
		"body:json {",
		"}",
	)
	block, found := file.Block("docs")
	if !found || block.Text != "# Title\n\n  indented\nno indent\n}" || block.TextLine != 2 || block.Line != 1 {
		t.Errorf("docs = %+v", block)
	}
	if empty, found := file.Block("body:json"); !found || empty.Text != "" || empty.TextLine != 9 {
		t.Errorf("empty = %+v", empty)
	}
}

func TestParseBruReadsListsAndListValues(t *testing.T) {
	file := parse(t,
		"vars:secret [",
		"  token,",
		"  ~old, other",
		"]",
		"meta {",
		"  tags: [",
		"    smoke,",
		"    regression",
		"  ]",
		"  inline: [a, b ,]",
		"  empty: []",
		"}",
		"color: #00ff00",
	)
	want := []dto.BruEntry{{Key: "token", Line: 2, Enabled: true}, {Key: "old", Line: 3, Enabled: false}, {Key: "other", Line: 3, Enabled: true}}
	if !reflect.DeepEqual(file.Entries("vars:secret"), want) {
		t.Errorf("secrets = %+v", file.Entries("vars:secret"))
	}
	meta, _ := file.Block("meta")
	tags, _ := meta.Lookup("tags")
	inline, _ := meta.Lookup("inline")
	empty, _ := meta.Lookup("empty")
	if !reflect.DeepEqual(tags.Values, []string{"smoke", "regression"}) || tags.Value != "smoke, regression" || tags.Line != 6 {
		t.Errorf("tags = %+v", tags)
	}
	if !reflect.DeepEqual(inline.Values, []string{"a", "b"}) || !reflect.DeepEqual(empty.Values, []string{}) {
		t.Errorf("inline = %+v empty = %+v", inline, empty)
	}
	if _, found := meta.Lookup("missing"); found {
		t.Error("missing key found")
	}
	if file.Attribute("color") != "#00ff00" || file.Attribute("missing") != "" {
		t.Errorf("attributes = %+v", file.Attributes)
	}
}

func TestParseBruReadsMultiLineValues(t *testing.T) {
	file := parse(t,
		"body:form-urlencoded {",
		"  note: '''",
		"    first",
		"",
		"      second",
		"  '''",
		"  after: x",
		"}",
	)
	entries := file.Entries("body:form-urlencoded")
	if len(entries) != 2 || entries[0].Value != "first\n\n  second" || entries[0].Line != 2 || entries[1].Key != "after" {
		t.Errorf("entries = %+v", entries)
	}
}

func TestParseBruAcceptsWindowsLineEndings(t *testing.T) {
	file, err := parseBru("a.bru", []byte("meta {\r\n  name: A\r\n}\r\nscript:pre-request {\r\n  a();\r\n  b();\r\n}\r\n"))
	if err != nil || file.Value("meta", "name") != "A" {
		t.Fatalf("file = %+v %v", file, err)
	}
	if block, _ := file.Block("script:pre-request"); block.Text != "a();\nb();" {
		t.Errorf("script = %q", block.Text)
	}
	if old, err := parseBru("a.bru", []byte("docs {\r  x\r}")); err != nil || old.Blocks[0].Text != "x" {
		t.Errorf("old mac = %+v %v", old, err)
	}
}

func TestParseBruRefusesBrokenFiles(t *testing.T) {
	cases := map[string]struct {
		text    string
		message string
	}{
		"unclosed text":      {"docs {\n  x", "Orders/a.bru:1 is not valid Bruno: the docs block is not closed"},
		"unclosed block":     {"\nmeta {\n  name: a", "Orders/a.bru:2 is not valid Bruno: the meta block is not closed"},
		"unclosed list":      {"vars:secret [\n  a", "Orders/a.bru:1 is not valid Bruno: the vars:secret list is not closed"},
		"unclosed multiline": {"vars {\n  a: '''\n  x", "Orders/a.bru:2 is not valid Bruno: the multi-line value of a is not closed"},
		"unclosed list item": {"meta {\n  tags: [\n  x", "Orders/a.bru:2 is not valid Bruno: the list tags is not closed"},
		"stray line":         {"meta {\n}\nnonsense", `Orders/a.bru:3 is not valid Bruno: unexpected line "nonsense"`},
		"bad entry":          {"meta {\n  just text\n}", `Orders/a.bru:2 is not valid Bruno: expected key: value, got "just text"`},
		"empty key":          {"meta {\n  : x\n}", `Orders/a.bru:2 is not valid Bruno: expected key: value, got ": x"`},
		"bad quoted key":     {"meta {\n  \"a: x\n}", `Orders/a.bru:2 is not valid Bruno: expected key: value, got "\"a: x"`},
		"quoted key no sep":  {"meta {\n  \"a\" x\n}", `Orders/a.bru:2 is not valid Bruno: expected key: value, got "\"a\" x"`},
	}
	for name, tc := range cases {
		_, err := parseBru("Orders/a.bru", []byte(tc.text))
		custom, isCustom := err.(errs.BaseError)
		if !isCustom || custom.Code() != "IMP-36400" || custom.Description() != tc.message {
			t.Errorf("%s: got %v", name, err)
		}
	}
}
