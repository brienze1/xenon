package entity

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/authtype"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/domain/enums/rawlanguage"
)

func TestIsValidItemName(t *testing.T) {
	cases := map[string]bool{
		"":                       false,
		"   ":                    false,
		"A":                      true,
		" Acme ":                 true,
		strings.Repeat("é", 200): true,
		strings.Repeat("a", 201): false,
	}
	for name, want := range cases {
		if got := isValidItemName(name); got != want {
			t.Errorf("isValidItemName(%d runes) = %v, want %v", len([]rune(name)), got, want)
		}
	}
}

func TestKeyValue(t *testing.T) {
	row := NewKeyValue("Accept", "application/json")
	if !row.Enabled || row.Key != "Accept" || row.Value != "application/json" || row.Description != "" {
		t.Fatalf("NewKeyValue = %+v", row)
	}
	if !row.IsActive() {
		t.Error("an enabled row with a key is active")
	}
	if (KeyValue{Key: "Accept"}).IsActive() {
		t.Error("a disabled row is not active")
	}
	if (KeyValue{Key: "  ", Enabled: true}).IsActive() {
		t.Error("a row without key is not active")
	}
	if !row.HasKey(" accept ") || row.HasKey("Accept-Encoding") {
		t.Error("HasKey must compare trimmed names case-insensitively")
	}
}

func TestPathVariableHasKey(t *testing.T) {
	variable := PathVariable{Key: "id", Value: "42"}
	if !variable.HasKey("id") || variable.HasKey("ID") {
		t.Error("path variable names are case-sensitive")
	}
}

func TestAuth(t *testing.T) {
	if auth := DefaultAuth(); auth.Type != authtype.Inherit || !auth.IsInherited() {
		t.Fatalf("DefaultAuth = %+v", auth)
	}
	if (Auth{Type: authtype.Bearer}).IsInherited() {
		t.Error("bearer is not inherited")
	}
	supported := map[string]bool{"inherit": true, "none": true, "bearer": true, "basic": true, "apikey": false, "oauth2": false, "awsv4": false, "digest": false}
	for raw, want := range supported {
		if got := (Auth{Type: authtype.GetAuthType(raw)}).IsSupported(); got != want {
			t.Errorf("IsSupported(%s) = %v, want %v", raw, got, want)
		}
	}
}

func TestFileRef(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "work", "acme")
	inside := NewFileRef(root, filepath.Join(root, "fixtures", "invoice.pdf"))
	if inside.Src != "fixtures/invoice.pdf" {
		t.Errorf("inside Src = %q", inside.Src)
	}
	sibling := filepath.Join(string(filepath.Separator), "work", "acme-other", "file.bin")
	if got := NewFileRef(root, sibling).Src; got != sibling {
		t.Errorf("sibling Src = %q, want absolute", got)
	}
	parent := filepath.Join(string(filepath.Separator), "work")
	if got := NewFileRef(root, parent).Src; got != parent {
		t.Errorf("parent Src = %q, want absolute", got)
	}
	if got := NewFileRef(root, "relative.txt").Src; got != "relative.txt" {
		t.Errorf("unrelated Src = %q", got)
	}
	if !(FileRef{Src: "  "}).IsEmpty() || inside.IsEmpty() {
		t.Error("IsEmpty misbehaves")
	}
	if got := inside.Resolve(root); got != filepath.Join(root, "fixtures", "invoice.pdf") {
		t.Errorf("Resolve(relative) = %q", got)
	}
	if got := (FileRef{Src: sibling}).Resolve(root); got != sibling {
		t.Errorf("Resolve(absolute) = %q", got)
	}
	if got := (FileRef{}).Resolve(root); got != "" {
		t.Errorf("Resolve(empty) = %q", got)
	}
}

func TestBody(t *testing.T) {
	body := DefaultBody()
	if body.Mode != bodymode.None || body.Raw.Language != rawlanguage.Json || body.FormData == nil || body.UrlEncoded == nil {
		t.Fatalf("DefaultBody = %+v", body)
	}
	if body.HasContent() {
		t.Error("mode none has no content")
	}
	body.Mode = bodymode.Raw
	if !body.HasContent() {
		t.Error("mode raw has content")
	}
	if refs := body.FileRefs(); len(refs) != 0 {
		t.Errorf("raw body refs = %v", refs)
	}
	binary := Body{Mode: bodymode.Binary, Binary: FileRef{Src: "payload.bin"}}
	if refs := binary.FileRefs(); len(refs) != 1 || refs[0].Src != "payload.bin" {
		t.Errorf("binary refs = %v", refs)
	}
	form := Body{Mode: bodymode.FormData, FormData: []FormDataRow{
		{Key: "file", Type: formdatatype.File, Src: FileRef{Src: "a.pdf"}, Enabled: true},
		{Key: "off", Type: formdatatype.File, Src: FileRef{Src: "b.pdf"}},
		{Key: "note", Type: formdatatype.Text, Value: "hello", Enabled: true},
	}}
	if refs := form.FileRefs(); len(refs) != 1 || refs[0].Src != "a.pdf" {
		t.Errorf("form refs = %v", refs)
	}
	if !form.FormData[0].IsFile() || form.FormData[2].IsFile() {
		t.Error("IsFile misbehaves")
	}
}

func TestRequestSettings(t *testing.T) {
	settings := DefaultRequestSettings()
	if settings.TimeoutMs != 0 || !settings.FollowRedirects || !settings.VerifySsl || !settings.EncodeUrl {
		t.Fatalf("DefaultRequestSettings = %+v", settings)
	}
	if settings.HasTimeout() {
		t.Error("0 means no timeout")
	}
	settings.TimeoutMs = 1
	if !settings.HasTimeout() {
		t.Error("1 ms is a timeout")
	}
	settings.TimeoutMs = 1500
	if settings.Timeout() != 1500*time.Millisecond {
		t.Errorf("Timeout() = %v", settings.Timeout())
	}
}
