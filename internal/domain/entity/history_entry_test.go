package entity

import (
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/historysource"
)

func TestInEnvironmentRecordsTheActiveEnvironmentAsItIsThen(t *testing.T) {
	active := NewEnvironment("hml", "HML")
	active.Color = "green"
	entry := HistoryEntry{ID: "h1"}.InEnvironment(&active)
	active.Name = "Renamed"
	if entry.Environment == nil || entry.Environment.ID != "hml" || entry.Environment.Name != "HML" || entry.Environment.Color != "green" {
		t.Fatalf("Environment = %+v, want the snapshot", entry.Environment)
	}
	if got := entry.InEnvironment(nil).Environment; got != nil {
		t.Errorf("no active environment records none, got %+v", got)
	}
}

func TestAIEntriesMatchTheirSessionTag(t *testing.T) {
	action := NewActionHistoryEntry("a", time.Time{}, "Claude", "AbC123", AIActivity{Tool: "tree", Action: "list"})
	send := NewHistoryEntry("s", time.Time{}, Request{Method: "GET", URL: "https://api.test/ping"}, SendExecution{}, AISendOrigin("Claude", "XyZ789"))
	user := NewHistoryEntry("u", time.Time{}, Request{Method: "GET", URL: "https://api.test/ping"}, SendExecution{}, SendOrigin{Source: historysource.User, Session: "XyZ789"})
	tests := []struct {
		name  string
		entry HistoryEntry
		query string
		want  bool
	}{
		{name: "action by tag", entry: action, query: "abc123", want: true},
		{name: "action by another tag", entry: action, query: "xyz789", want: false},
		{name: "send by tag", entry: send, query: " XYZ7 ", want: true},
		{name: "send keeps matching its url", entry: send, query: "ping", want: true},
		{name: "a user send has no tag", entry: user, query: "xyz789", want: false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.entry.Matches(test.query); got != test.want {
				t.Errorf("Matches(%q) = %v, want %v", test.query, got, test.want)
			}
		})
	}
	if send.Session != "XyZ789" || user.Session != "" {
		t.Errorf("sessions = %q, %q", send.Session, user.Session)
	}
	if normalized := action.Normalized(); normalized.Session != "AbC123" {
		t.Errorf("Normalized() dropped the session: %+v", normalized)
	}
}
