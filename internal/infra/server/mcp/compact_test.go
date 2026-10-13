package mcp

import (
	"errors"
	"strings"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

func codeOf(err error) string {
	var custom errs.BaseError
	if errors.As(err, &custom) {
		return custom.Code()
	}
	return ""
}

func TestPaginateWalksPagesWithAnOpaqueCursor(t *testing.T) {
	items := make([]int, 250)
	for index := range items {
		items[index] = index
	}
	first, err := Paginate("tree", items, "")
	if err != nil || len(first.Items) != PageSize || first.Next == "" || first.Items[0] != 0 {
		t.Fatalf("first page = %d items, next %q, %v", len(first.Items), first.Next, err)
	}
	second, _ := Paginate("tree", items, first.Next)
	third, _ := Paginate("tree", items, second.Next)
	if second.Items[0] != 100 || len(third.Items) != 50 || third.Next != "" {
		t.Fatalf("second starts at %d, third has %d items and next %q", second.Items[0], len(third.Items), third.Next)
	}
	beyond, _ := Paginate("tree", items[:3], encodeCursor(10))
	if len(beyond.Items) != 0 || beyond.Items == nil {
		t.Errorf("a cursor past the end must give an empty page: %+v", beyond)
	}
	for _, cursor := range []string{"!!", "Zm9v", "bzpi"} {
		if _, err := Paginate("tree", items, cursor); codeOf(err) != "MCP-07400" {
			t.Errorf("cursor %q: %v", cursor, err)
		}
	}
}

func TestInlineAndSlicedBodiesStayOnRuneBoundaries(t *testing.T) {
	short, truncated, size := InlineBody("hello")
	if short != "hello" || truncated || size != 5 {
		t.Fatalf("short body = %q %v %d", short, truncated, size)
	}
	long := strings.Repeat("a", InlineBodyBytes-1) + "é" + strings.Repeat("b", 10)
	inline, truncated, size := InlineBody(long)
	if !truncated || size != len(long) || len(inline) != InlineBodyBytes-1 {
		t.Fatalf("long body cut at %d bytes, truncated %v, size %d", len(inline), truncated, size)
	}
	slice, err := SliceBody("response", long, InlineBodyBytes, 0)
	if err != nil || slice.Offset != InlineBodyBytes-1 || !strings.HasPrefix(slice.Text, "é") || slice.Truncated || slice.Size != len(long) {
		t.Fatalf("slice = %+v, %v", slice, err)
	}
	for _, bad := range [][2]int{{-1, 10}, {0, MaxBodyReadBytes + 1}, {0, -1}} {
		if _, err := SliceBody("response", long, bad[0], bad[1]); codeOf(err) != "MCP-07400" {
			t.Errorf("SliceBody(%v) = %v", bad, err)
		}
	}
	if past, _ := SliceBody("response", "abc", 10, 5); past.Text != "" || past.Offset != 3 {
		t.Errorf("offset past the end = %+v", past)
	}
}

func TestEncodeJSONOmitsEmptyFieldsButKeepsItems(t *testing.T) {
	type row struct {
		Name   string   `json:"name"`
		Note   string   `json:"note"`
		Passed bool     `json:"passed"`
		Tags   []string `json:"tags"`
	}
	payload := map[string]any{"items": []row{}, "extra": map[string]any{}, "html": "<a&b>", "rows": []row{{Name: "x"}}, "pair": []string{"k", ""}, "nothing": nil}
	encoded, err := encodeJSON(payload)
	if err != nil || encoded != `{"html":"<a&b>","items":[],"pair":["k",""],"rows":[{"name":"x","passed":false}]}` {
		t.Fatalf("encodeJSON() = %s, %v", encoded, err)
	}
	if empty, _ := encodeJSON(map[string]any{}); empty != `{}` {
		t.Errorf("an empty result = %s", empty)
	}
	if _, err := encodeJSON(func() {}); err == nil {
		t.Error("an unencodable payload must fail")
	}
}

func TestMcpErrorsCarryTheirCodes(t *testing.T) {
	cases := map[string]error{
		"MCP-07400": invalidArguments("tree", "bad"),
		"MCP-06404": noActiveWorkspace(),
		"MCP-02403": humanOnly("Revealing a secret"),
		"MCP-08404": itemNotFound("01X"),
	}
	for code, err := range cases {
		if codeOf(err) != code {
			t.Errorf("%s: got %v", code, err)
		}
	}
	if call := (Call{Tool: "tree", Action: "list", Session: "s1"}); codeOf(call.Unsupported()) != "MCP-07400" || call.TabId() != "mcp:s1" {
		t.Errorf("call helpers: %v %s", call.Unsupported(), call.TabId())
	}
	var target struct {
		Name string `json:"name"`
	}
	if err := (Call{Tool: "tree", Arguments: []byte(`{"name":"x"}`)}).Decode(&target); err != nil || target.Name != "x" {
		t.Errorf("Decode() = %v %+v", err, target)
	}
	if err := (Call{Tool: "tree", Arguments: []byte(`{"name":1}`)}).Decode(&target); codeOf(err) != "MCP-07400" {
		t.Errorf("Decode() of a wrong type = %v", err)
	}
	if err := (Call{Tool: "tree"}).Decode(&target); err != nil {
		t.Errorf("Decode() of no arguments = %v", err)
	}
}
