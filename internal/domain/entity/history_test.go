package entity

import (
	"strconv"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

func historyOf(count int) []HistoryEntry {
	entries := make([]HistoryEntry, 0, count)
	for index := 1; index <= count; index++ {
		request := NewRequest("", "r"+strconv.Itoa(index))
		request.URL = "https://api.test/item/" + strconv.Itoa(index)
		entries = append(entries, HistoryEntry{ID: strconv.Itoa(index), Request: request})
	}
	return entries
}

func ids(entries []HistoryEntry) string {
	result := ""
	for _, entry := range entries {
		result += entry.ID + ","
	}
	return result
}

func TestNewHistoryEntryKeepsTheRequestAsWrittenAndTheResult(t *testing.T) {
	request := NewRequest("01JREQ", "Ping")
	request.Method = "purge"
	request.URL = "{{base}}/Ping"
	sentAt := time.Date(2026, 10, 9, 10, 0, 0, 0, time.UTC)
	ok := NewResponse(204, "No Content", "HTTP/1.1", nil, nil, 0, Timings{Total: 15 * time.Millisecond})
	execution := NewSendExecution("01JSND")
	execution.RecordResponse(EffectiveRequest{}, ok)
	entry := NewHistoryEntry("01JHIS", sentAt, request, execution, UserSendOrigin())
	if entry.ID != "01JHIS" || entry.SourceRequestID != "01JREQ" || !entry.SentAt.Equal(sentAt) || entry.Request.URL != "{{base}}/Ping" {
		t.Fatalf("NewHistoryEntry = %+v", entry)
	}
	if !entry.Result.IsOk() || entry.Result.StatusCode != 204 || entry.Result.StatusText != "No Content" || entry.Result.Total != 15*time.Millisecond {
		t.Fatalf("result = %+v", entry.Result)
	}
	execution.RecordResponse(EffectiveRequest{}, NewFailedResponse(NewSendError(senderrorkind.Timeout, "10"), Timings{}))
	failed := NewHistoryEntry("01JHIT", sentAt, request, execution, UserSendOrigin())
	if failed.Result.IsOk() || failed.Result.ErrorKind != senderrorkind.Timeout || failed.Result.ErrorMessage != "Timed out after 10 ms" {
		t.Fatalf("failed result = %+v", failed.Result)
	}
	for query, want := range map[string]bool{"PURGE": true, " ping ": true, "{{BASE}}": true, "get": false} {
		if entry.Matches(query) != want {
			t.Errorf("Matches(%q) = %v, want %v", query, !want, want)
		}
	}
	if logged := entry.LogValue().(map[string]any); logged["id"] != "01JHIS" || logged["statusCode"] != 204 {
		t.Errorf("LogValue = %v", logged)
	}
}

func TestHistoryKeepsTheNewestEntriesNewestFirst(t *testing.T) {
	history := NewHistory(historyOf(DefaultHistoryCap+2), DefaultHistoryCaps)
	page := history.Page(NewHistoryQuery("", "", 0))
	if len(page) != DefaultHistoryCap || page[0].ID != strconv.Itoa(DefaultHistoryCap+2) || page[DefaultHistoryCap-1].ID != "3" {
		t.Fatalf("page has %d entries from %s to %s", len(page), page[0].ID, page[len(page)-1].ID)
	}
	if _, found := history.Find("2"); found {
		t.Error("entries beyond the cap must be dropped")
	}
	if entry, found := history.Find("3"); !found || entry.ID != "3" {
		t.Error("the oldest kept entry must be found")
	}
	if got := ids(NewHistory(historyOf(3), DefaultHistoryCaps).Page(NewHistoryQuery("", "", 0))); got != "3,2,1," {
		t.Errorf("small history page = %s", got)
	}
}

func TestHistoryPagesWithQueryLimitAndCursor(t *testing.T) {
	history := NewHistory(historyOf(12), DefaultHistoryCaps)
	cases := []struct {
		query HistoryQuery
		want  string
	}{
		{NewHistoryQuery("", "", 2), "12,11,"},
		{NewHistoryQuery("", "11", 2), "10,9,"},
		{NewHistoryQuery("", "2", 5), "1,"},
		{NewHistoryQuery("", "1", 5), ""},
		{NewHistoryQuery("", "unknown", 5), ""},
		{NewHistoryQuery("item/1", "", 0), "12,11,10,1,"},
		{NewHistoryQuery("item/1", "11", 2), "10,1,"},
		{NewHistoryQuery("delete", "", 0), ""},
	}
	for _, test := range cases {
		if got := ids(history.Page(test.query)); got != test.want {
			t.Errorf("Page(%+v) = %s, want %s", test.query, got, test.want)
		}
	}
	if NewHistoryQuery("x", "y", 7) != (HistoryQuery{Source: historyfilter.All, Text: "x", BeforeID: "y", Limit: 7}) {
		t.Error("NewHistoryQuery must keep explicit values")
	}
}
