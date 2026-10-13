package entity

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestSecretFormsCoverEveryEncoding(t *testing.T) {
	forms := SecretForms(`a b/"c"<d>`)
	want := []string{`a b/"c"<d>`, `a+b%2F%22c%22%3Cd%3E`, `a%20b%2F%22c%22%3Cd%3E`, `a%20b%2F%22c%22%3Cd%3E`, `a b/\"c\"<d>`, `a b/\"c\"<d>`, base64.StdEncoding.EncodeToString([]byte(`a b/"c"<d>`))}
	for _, form := range want {
		if !containsString(forms, form) {
			t.Errorf("forms %q miss %q", forms, form)
		}
	}
	if len(SecretForms("")) != 0 {
		t.Error("an empty value has no forms")
	}
}

func TestKnownSecretsMask(t *testing.T) {
	known := NewKnownSecrets("s3cr3t-token", "abc", "", "s3cr3t", "p@ss word")
	cases := []struct {
		name string
		text string
		want string
	}{
		{name: "raw", text: `{"auth":"Bearer s3cr3t-token"}`, want: `{"auth":"Bearer •••••"}`},
		{name: "longest first", text: `s3cr3t-token and s3cr3t`, want: `••••• and •••••`},
		{name: "url encoded", text: `?q=p%40ss+word&r=p%40ss%20word`, want: `?q=•••••&r=•••••`},
		{name: "base64", text: base64.StdEncoding.EncodeToString([]byte("s3cr3t-token")), want: `•••••`},
		{name: "short whole value", text: `{"v":"abc"}`, want: `{"v":"•••••"}`},
		{name: "short inside words", text: `{"v":"abcdef"}`, want: `{"v":"abcdef"}`},
		{name: "nothing known", text: `plain`, want: `plain`},
	}
	for _, test := range cases {
		if got := known.Mask(test.text); got != test.want {
			t.Errorf("%s: Mask(%q) = %q, want %q", test.name, test.text, got, test.want)
		}
	}
	if known.Count() != 4 {
		t.Errorf("Count() = %d", known.Count())
	}
	grown := known.With("another")
	if grown.Count() != 5 || known.Count() != 4 {
		t.Errorf("With must not change the original: %d %d", grown.Count(), known.Count())
	}
	var empty KnownSecrets
	empty.Add("late")
	if empty.Count() != 1 || strings.Join(empty.Values(), ",") != "late" {
		t.Errorf("zero value Add = %v", empty.Values())
	}
	if logged := known.LogValue().(map[string]any); len(logged) != 1 || logged["count"] != 4 {
		t.Errorf("LogValue() = %v", logged)
	}
	if forms := known.Forms(); len(forms[0]) < len(forms[len(forms)-1]) {
		t.Errorf("forms are not longest-first: %q", forms)
	}
}

func TestVaultKnownSecretValues(t *testing.T) {
	vault := Vault{State: "unlocked", Secrets: map[string]map[string]string{"globals": {"a": "1"}}, OAuth2Tokens: map[string]OAuth2Token{"o": {AccessToken: "tok"}}}
	if values := vault.KnownSecretValues(); len(values) != 2 {
		t.Errorf("unlocked values = %v", values)
	}
	vault.State = "locked"
	if values := vault.KnownSecretValues(); len(values) != 0 {
		t.Errorf("locked values = %v", values)
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func TestSendExecutionDerivesTheBasicAuthCredentials(t *testing.T) {
	execution := SendExecution{SentRequest: &EffectiveRequest{Auth: Auth{Type: "basic", Basic: AuthBasic{Username: "user", Password: "s3cret"}}}}
	derived := execution.DerivedSecrets([]string{"s3cret"})
	if len(derived) != 1 || derived[0] != base64.StdEncoding.EncodeToString([]byte("user:s3cret")) {
		t.Fatalf("derived = %v", derived)
	}
	if derived := execution.DerivedSecrets([]string{"other"}); len(derived) != 0 {
		t.Errorf("no secret took part: %v", derived)
	}
	if derived := (SendExecution{}).DerivedSecrets([]string{"s3cret"}); len(derived) != 0 {
		t.Errorf("nothing was sent: %v", derived)
	}
}
