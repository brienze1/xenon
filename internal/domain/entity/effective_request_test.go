package entity

import (
	"reflect"
	"testing"
)

func TestEffectiveRequestHeaders(t *testing.T) {
	request := EffectiveRequest{
		Method:              "GET",
		URL:                 "https://h",
		Headers:             []KeyValue{NewKeyValue("X-Token", "abc")},
		DisabledAutoHeaders: []string{" accept "},
		Settings:            RequestSettings{TimeoutMs: 250},
	}
	if !request.HasHeader("x-token") || request.HasHeader("Accept") {
		t.Fatal("HasHeader must match user headers by name, case-insensitively")
	}
	if !request.IsAutoHeaderDisabled("Accept") || request.IsAutoHeaderDisabled("User-Agent") {
		t.Fatal("IsAutoHeaderDisabled must match trimmed names case-insensitively")
	}
	if request.TimeoutText() != "250" {
		t.Fatalf("TimeoutText = %q", request.TimeoutText())
	}
	if got := request.LogValue(); !reflect.DeepEqual(got, map[string]any{"method": "GET", "url": "https://h"}) {
		t.Fatalf("LogValue = %v", got)
	}

	withAuto := request.WithAutoHeaders([]AutoHeader{
		NewComputedAutoHeader(HeaderContentType, "multipart"),
		{Key: HeaderAccept, Value: "*/*", Disabled: true},
		NewAutoHeader(HeaderUserAgent, "Xenon/1"),
	})
	if request.AutoHeaders != nil {
		t.Fatal("WithAutoHeaders must not change the receiver")
	}
	if !withAuto.UsesAutoHeader("content-type") || withAuto.UsesAutoHeader("Accept") || withAuto.UsesAutoHeader("Host") {
		t.Fatal("UsesAutoHeader must be true only for listed, enabled auto headers")
	}
	want := []KeyValue{NewKeyValue("X-Token", "abc"), NewKeyValue(HeaderUserAgent, "Xenon/1")}
	if got := withAuto.SendHeaders(); !reflect.DeepEqual(got, want) {
		t.Fatalf("SendHeaders = %+v, want %+v", got, want)
	}
	if len(request.Headers) != 1 {
		t.Fatal("SendHeaders must not change the user headers")
	}
}

func TestAutoHeader(t *testing.T) {
	plain := NewAutoHeader("A", "1")
	computed := NewComputedAutoHeader("B", "2")
	disabled := AutoHeader{Key: "C", Disabled: true}
	if plain != (AutoHeader{Key: "A", Value: "1"}) || computed != (AutoHeader{Key: "B", Value: "2", Computed: true}) {
		t.Fatalf("constructors: %+v %+v", plain, computed)
	}
	if !plain.IsSent() || computed.IsSent() || disabled.IsSent() {
		t.Fatal("only enabled, non-computed auto headers are sent")
	}
	if !plain.IsActive() || !computed.IsActive() || disabled.IsActive() {
		t.Fatal("only disabled auto headers are inactive")
	}
}
