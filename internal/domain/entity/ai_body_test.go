package entity

import (
	"strings"
	"testing"
)

func TestInlineAndSlicedAIBodiesStayOnRuneBoundaries(t *testing.T) {
	short := NewAIInlineBody("hello")
	if short.Text != "hello" || short.Truncated || short.Size != 5 {
		t.Fatalf("short body = %+v", short)
	}
	long := strings.Repeat("a", AIInlineBodyBytes-1) + "é" + strings.Repeat("b", 10)
	inline := NewAIInlineBody(long)
	if !inline.Truncated || inline.Size != len(long) || len(inline.Text) != AIInlineBodyBytes-1 {
		t.Fatalf("long body cut at %d bytes, truncated %v, size %d", len(inline.Text), inline.Truncated, inline.Size)
	}
	slice, valid := NewAIBodySlice(long, AIInlineBodyBytes, 0)
	if !valid || slice.Offset != AIInlineBodyBytes-1 || !strings.HasPrefix(slice.Text, "é") || slice.Truncated || slice.Size != len(long) {
		t.Fatalf("slice = %+v, %v", slice, valid)
	}
	for _, bad := range [][2]int{{-1, 10}, {0, AIMaxBodyReadBytes + 1}, {0, -1}} {
		if _, valid := NewAIBodySlice(long, bad[0], bad[1]); valid {
			t.Errorf("NewAIBodySlice(%v) must be rejected", bad)
		}
	}
	if past, _ := NewAIBodySlice("abc", 10, 5); past.Text != "" || past.Offset != 3 {
		t.Errorf("offset past the end = %+v", past)
	}
	if part, _ := NewAIBodySlice("abcdef", 1, 2); part.Text != "bc" || !part.Truncated {
		t.Errorf("a partial slice = %+v", part)
	}
}

func TestIsAIBinaryBody(t *testing.T) {
	cases := []struct {
		binary      bool
		contentType string
		want        bool
	}{
		{binary: true, contentType: "text/plain", want: true},
		{contentType: "application/octet-stream", want: true},
		{contentType: "application/octet-stream; charset=binary", want: true},
		{contentType: "application/json"},
		{contentType: ""},
	}
	for _, tc := range cases {
		if got := IsAIBinaryBody(tc.binary, tc.contentType); got != tc.want {
			t.Errorf("IsAIBinaryBody(%v, %q) = %v", tc.binary, tc.contentType, got)
		}
	}
}

func TestAISessionTabID(t *testing.T) {
	if AISessionTabID("s1") != "mcp:s1" {
		t.Errorf("AISessionTabID() = %q", AISessionTabID("s1"))
	}
}
