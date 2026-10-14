package entity

import "testing"

func TestNewAIPageWalksPagesWithAnOpaqueCursor(t *testing.T) {
	items := make([]int, 250)
	for index := range items {
		items[index] = index
	}
	first, valid := NewAIPage(items, "")
	if !valid || len(first.Items) != AIPageSize || first.Next == "" || first.Items[0] != 0 {
		t.Fatalf("first page = %d items, next %q, %v", len(first.Items), first.Next, valid)
	}
	second, _ := NewAIPage(items, first.Next)
	third, _ := NewAIPage(items, second.Next)
	if second.Items[0] != 100 || len(third.Items) != 50 || third.Next != "" {
		t.Fatalf("second starts at %d, third has %d items and next %q", second.Items[0], len(third.Items), third.Next)
	}
	beyond, _ := NewAIPage(items[:3], encodeAICursor(10))
	if len(beyond.Items) != 0 || beyond.Items == nil {
		t.Errorf("a cursor past the end must give an empty page: %+v", beyond)
	}
	for _, cursor := range []string{"!!", "Zm9v", "bzpi"} {
		if _, valid := NewAIPage(items, cursor); valid {
			t.Errorf("cursor %q must be unknown", cursor)
		}
	}
}

func TestAICursorRoundTripsAndKeepsItsWireForm(t *testing.T) {
	if encodeAICursor(100) != "bzoxMDA" {
		t.Errorf("encodeAICursor(100) = %q", encodeAICursor(100))
	}
	cases := []struct {
		raw    string
		offset int
		valid  bool
	}{
		{raw: "", offset: 0, valid: true},
		{raw: encodeAICursor(42), offset: 42, valid: true},
		{raw: "!!"},
		{raw: "Zm9v"},
		{raw: "bzpi"},
		{raw: encodeAICursor(-1)},
	}
	for _, tc := range cases {
		offset, valid := ParseAICursor(tc.raw)
		if offset != tc.offset || valid != tc.valid {
			t.Errorf("ParseAICursor(%q) = %d %v, want %d %v", tc.raw, offset, valid, tc.offset, tc.valid)
		}
	}
}
