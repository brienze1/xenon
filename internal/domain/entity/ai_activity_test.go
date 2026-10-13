package entity

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func TestSendOriginNamesTheClient(t *testing.T) {
	cases := []struct {
		origin SendOrigin
		source string
		client string
	}{
		{origin: UserSendOrigin(), source: "user", client: ""},
		{origin: AISendOrigin(" Claude Code "), source: "ai", client: "Claude Code"},
		{origin: AISendOrigin(""), source: "ai", client: DefaultAIClientName},
		{origin: SendOrigin{}.Normalized(), source: "user", client: ""},
		{origin: SendOrigin{Source: "ai"}.Normalized(), source: "ai", client: DefaultAIClientName},
	}
	for _, test := range cases {
		if test.origin.Source.String() != test.source || test.origin.Client != test.client {
			t.Errorf("origin = %+v, want %s/%s", test.origin, test.source, test.client)
		}
	}
	if !AISendOrigin("x").IsAI() || UserSendOrigin().IsAI() {
		t.Error("IsAI")
	}
	if logged := AISendOrigin("x").LogValue().(map[string]any); logged["client"] != "x" {
		t.Errorf("LogValue() = %v", logged)
	}
}

func TestAIActivityArgsAreCutAt2KB(t *testing.T) {
	cases := []struct {
		name  string
		args  string
		bytes int
	}{
		{name: "short", args: `{"action":"list"}`, bytes: 17},
		{name: "exactly 2 KB", args: strings.Repeat("x", MaxAIActivityArgsBytes), bytes: MaxAIActivityArgsBytes},
		{name: "longer", args: strings.Repeat("x", 3000), bytes: MaxAIActivityArgsBytes + len("…")},
		{name: "rune boundary", args: "x" + strings.Repeat("é", 2000), bytes: MaxAIActivityArgsBytes - 1 + len("…")},
	}
	for _, test := range cases {
		got := AIActivity{}.WithArgs(test.args).Args
		if len(got) != test.bytes || !utf8.ValidString(got) {
			t.Errorf("%s: %d bytes, want %d", test.name, len(got), test.bytes)
		}
	}
}

func TestActionHistoryEntriesMatchTheirActivity(t *testing.T) {
	at := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	entry := NewActionHistoryEntry("01", at, "", AIActivity{Tool: "tree_edit", Action: "create", Target: AIActivityTarget{Name: "Payments"}, Outcome: AIActivityOk})
	if !entry.IsAction() || entry.Client != DefaultAIClientName || entry.Origin() != "ai" || !entry.IsIn("ai") || entry.IsIn("user") || !entry.IsIn("all") {
		t.Fatalf("entry = %+v", entry)
	}
	for query, want := range map[string]bool{"TREE_EDIT": true, "create": true, "paym": true, "list": false} {
		if entry.Matches(query) != want {
			t.Errorf("Matches(%q) = %v", query, !want)
		}
	}
	if logged := entry.LogValue().(map[string]any); logged["kind"] != entry.Kind {
		t.Errorf("LogValue() = %v", logged)
	}
	if logged := (AIActivity{Tool: "x"}).LogValue().(map[string]any); logged["tool"] != "x" {
		t.Errorf("activity LogValue() = %v", logged)
	}
	history := NewHistory([]HistoryEntry{{ID: "old"}, entry}, DefaultHistoryCaps)
	if got := ids(history.Page(NewHistoryQuery("", "", 0).FromSource("user"))); got != "old," {
		t.Errorf("user page = %s", got)
	}
	if got := ids(history.Page(NewHistoryQuery("", "", 0).FromSource("ai"))); got != "01," {
		t.Errorf("ai page = %s", got)
	}
	if kept := history.WithoutSource("ai"); len(kept) != 1 || kept[0].ID != "old" {
		t.Errorf("WithoutSource(ai) = %v", kept)
	}
}

func TestHistoryKeepsTheNewestEntriesOfEachSource(t *testing.T) {
	entries := make([]HistoryEntry, 0, 250)
	for index := 0; index < 150; index++ {
		entries = append(entries, HistoryEntry{ID: "ai", Source: "ai"})
	}
	for index := 0; index < 100; index++ {
		entries = append(entries, HistoryEntry{ID: "user", Source: "user"})
	}
	history := NewHistory(entries, HistoryCaps{User: 1000, AI: 100})
	if page := history.Page(NewHistoryQuery("", "", 1000)); len(page) != 200 {
		t.Fatalf("kept %d entries, want 200", len(page))
	}
	if caps := (HistoryCaps{User: 5, AI: 30000}).OrDefault(); caps != DefaultHistoryCaps {
		t.Errorf("OrDefault() = %+v", caps)
	}
}
