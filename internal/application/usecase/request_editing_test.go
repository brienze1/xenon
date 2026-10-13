package usecase

import (
	"reflect"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
)

func disabledRow(key, value, description string) entity.KeyValue {
	return entity.KeyValue{Key: key, Value: value, Description: description}
}

func describedRow(key, value, description string) entity.KeyValue {
	row := entity.NewKeyValue(key, value)
	row.Description = description
	return row
}

func TestDeriveParamsFromUrl(t *testing.T) {
	cases := map[string]struct {
		url      string
		existing []entity.KeyValue
		want     []entity.KeyValue
	}{
		"fresh query": {
			url:  "https://h/a?page=2&limit=10",
			want: []entity.KeyValue{entity.NewKeyValue("page", "2"), entity.NewKeyValue("limit", "10")},
		},
		"keeps disabled rows and descriptions in place": {
			url:      "https://h/a?page=3&limit=10",
			existing: []entity.KeyValue{describedRow("page", "2", "Page"), disabledRow("debug", "1", "dev"), entity.NewKeyValue("limit", "9")},
			want:     []entity.KeyValue{describedRow("page", "3", "Page"), disabledRow("debug", "1", "dev"), entity.NewKeyValue("limit", "10")},
		},
		"drops enabled rows that left the URL": {
			url:      "https://h/a?page=2",
			existing: []entity.KeyValue{entity.NewKeyValue("page", "2"), disabledRow("debug", "1", ""), describedRow("limit", "10", "gone")},
			want:     []entity.KeyValue{entity.NewKeyValue("page", "2"), disabledRow("debug", "1", "")},
		},
		"appends new pairs": {
			url:      "https://h/a?a=1&b=2&c=3",
			existing: []entity.KeyValue{entity.NewKeyValue("a", "1"), disabledRow("x", "9", "")},
			want:     []entity.KeyValue{entity.NewKeyValue("a", "1"), disabledRow("x", "9", ""), entity.NewKeyValue("b", "2"), entity.NewKeyValue("c", "3")},
		},
		"no query keeps disabled rows only": {
			url:      "https://h/a",
			existing: []entity.KeyValue{entity.NewKeyValue("a", "1"), disabledRow("x", "9", "")},
			want:     []entity.KeyValue{disabledRow("x", "9", "")},
		},
	}
	for name, tc := range cases {
		request := entity.NewRequest("", "r")
		request.URL = tc.url
		request.Params = tc.existing
		if got := DeriveParamsFromUrl(request).Params; !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s: params = %+v, want %+v", name, got, tc.want)
		}
	}
}

func TestDeriveParamsFromUrlPathVariables(t *testing.T) {
	request := entity.NewRequest("", "r")
	request.URL = "https://h/customers/:customerId/invoices/:invoiceId"
	request.PathVariables = []entity.PathVariable{{Key: "invoiceId", Value: "7", Description: "inv"}, {Key: "gone", Value: "1"}}
	got := DeriveParamsFromUrl(request).PathVariables
	want := []entity.PathVariable{{Key: "customerId"}, {Key: "invoiceId", Value: "7", Description: "inv"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("path variables = %+v, want %+v", got, want)
	}
}

func TestBuildUrlFromParams(t *testing.T) {
	request := entity.NewRequest("", "r")
	request.URL = "https://h/users/:id/posts/:postId?old=1#top"
	request.Params = []entity.KeyValue{entity.NewKeyValue("page", "2"), disabledRow("debug", "1", ""), entity.NewKeyValue("q", "a b")}
	request.PathVariables = []entity.PathVariable{{Key: "userId", Value: "1"}, {Key: "postId"}}
	built := BuildUrlFromParams(request)
	if built.URL != "https://h/users/:userId/posts/:postId?page=2&q=a b#top" {
		t.Fatalf("URL = %q", built.URL)
	}
	if !reflect.DeepEqual(built.Params, request.Params) || !reflect.DeepEqual(built.PathVariables, request.PathVariables) {
		t.Fatal("the tables are returned unchanged")
	}
}

func TestParseBulkRows(t *testing.T) {
	got := ParseBulkRows("\n  page : 2  \r\n//debug:1\n// verbose : yes\n//\n   \nflag\nurl:https://h:8443/a\n:only value")
	want := []entity.KeyValue{
		entity.NewKeyValue("page", "2"),
		disabledRow("debug", "1", ""),
		disabledRow("verbose", "yes", ""),
		entity.NewKeyValue("flag", ""),
		entity.NewKeyValue("url", "https://h:8443/a"),
		entity.NewKeyValue("", "only value"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseBulkRows = %+v, want %+v", got, want)
	}
	if rows := ParseBulkRows(""); rows == nil || len(rows) != 0 {
		t.Fatalf("empty text gives no rows, got %#v", rows)
	}
}

func TestFormatBulkRows(t *testing.T) {
	rows := []entity.KeyValue{
		describedRow("page", "2", "dropped"),
		disabledRow("debug", "1", ""),
		entity.NewKeyValue("url", "a:b"),
		{Enabled: true},
		entity.NewKeyValue("flag", ""),
		entity.NewKeyValue("", "value only"),
	}
	if got := FormatBulkRows(rows); got != "page:2\n//debug:1\nurl:a:b\nflag:\n:value only" {
		t.Fatalf("FormatBulkRows = %q", got)
	}
	if got := FormatBulkRows(nil); got != "" {
		t.Fatalf("no rows = %q", got)
	}
}
