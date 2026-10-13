package sandbox

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/brienze1/xenon/internal/domain/enums/scriptmemberkind"
	"github.com/brienze1/xenon/internal/domain/enums/snippetphase"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestCheckScript(t *testing.T) {
	tools := NewScriptToolsWebservice()
	ctx := context.Background()
	if errors := tools.Check(ctx, "const a = await Promise.resolve(1)\nreturn a"); len(errors) != 0 {
		t.Fatalf("valid script errors = %+v", errors)
	}
	errors := tools.Check(ctx, "let ok = 1\nconst a = ;")
	if len(errors) != 1 || errors[0].Message != "SyntaxError: Unexpected token ';'" || errors[0].Line != 2 || errors[0].Column != 11 || errors[0].Script.Text != "let ok = 1\nconst a = ;" {
		t.Fatalf("errors = %+v", errors)
	}
	unclosed := tools.Check(ctx, "ok()\nif (x) {")
	if len(unclosed) != 1 || unclosed[0].Line != 2 || unclosed[0].Column != 9 {
		t.Fatalf("unclosed = %+v", unclosed)
	}
}

func TestWithinScript(t *testing.T) {
	cases := []struct {
		text                               string
		line, column, wantLine, wantColumn int
	}{
		{text: "a\nbc", line: 2, column: 1, wantLine: 2, wantColumn: 1},
		{text: "a\nbc", line: 3, column: 1, wantLine: 2, wantColumn: 3},
		{text: "é", line: 2, column: 5, wantLine: 1, wantColumn: 2},
		{text: "", line: 1, column: 4, wantLine: 1, wantColumn: 4},
	}
	for _, tc := range cases {
		if line, column := withinScript(tc.text, tc.line, tc.column); line != tc.wantLine || column != tc.wantColumn {
			t.Errorf("withinScript(%q, %d, %d) = %d:%d", tc.text, tc.line, tc.column, line, column)
		}
	}
}

func TestScriptReferenceAndSnippets(t *testing.T) {
	ctx := context.Background()
	tools := NewScriptToolsWebservice()
	members, err := tools.FindReference(ctx)
	if err != nil || len(members) < 200 || members[0].Path != "pm" || members[0].Kind != scriptmemberkind.Object || members[0].Description == "" {
		t.Fatalf("members = %d %+v, %v", len(members), members[0], err)
	}
	snippets, err := tools.FindSnippets(ctx)
	if err != nil || len(snippets) != 21 || snippets[0].ID != "get-environment-variable" || snippets[0].Phase != snippetphase.Any || snippets[11].Phase != snippetphase.PostResponse || snippets[11].Title != "Status code: Code is 200" {
		t.Fatalf("snippets = %+v, %v", snippets, err)
	}
}

func TestScriptToolsDataErrors(t *testing.T) {
	ctx := context.Background()
	missing := &scriptToolsWebservice{files: fstest.MapFS{}}
	if _, err := missing.FindReference(ctx); errorCodeOf(err) != "SCR-06500" {
		t.Errorf("missing reference = %v", err)
	}
	if _, err := missing.FindSnippets(ctx); errorCodeOf(err) != "SCR-08500" {
		t.Errorf("missing snippets = %v", err)
	}
	broken := &scriptToolsWebservice{files: fstest.MapFS{referenceFile: {Data: []byte("{")}, snippetsFile: {Data: []byte("[")}}}
	if _, err := broken.FindReference(ctx); errorCodeOf(err) != "SCR-07500" {
		t.Errorf("broken reference = %v", err)
	}
	if _, err := broken.FindSnippets(ctx); errorCodeOf(err) != "SCR-09500" {
		t.Errorf("broken snippets = %v", err)
	}
}

func errorCodeOf(err error) string {
	base, isBase := err.(errs.BaseError)
	if !isBase {
		return ""
	}
	return base.Code()
}
