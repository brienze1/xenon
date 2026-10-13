package entity

import (
	"bytes"
	"cmp"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"slices"
	"strings"
)

const (
	MinSecretMaskLength = 4
	jsonQuote           = `"`
)

type KnownSecrets struct {
	values map[string]struct{}
}

func NewKnownSecrets(values ...string) KnownSecrets {
	known := KnownSecrets{values: map[string]struct{}{}}
	known.Add(values...)
	return known
}

func (k *KnownSecrets) Add(values ...string) {
	if k.values == nil {
		k.values = map[string]struct{}{}
	}
	for _, value := range values {
		if value != "" {
			k.values[value] = struct{}{}
		}
	}
}

func (k KnownSecrets) With(values ...string) KnownSecrets {
	union := NewKnownSecrets(k.Values()...)
	union.Add(values...)
	return union
}

func (k KnownSecrets) Values() []string {
	values := make([]string, 0, len(k.values))
	for value := range k.values {
		values = append(values, value)
	}
	slices.Sort(values)
	return values
}

func (k KnownSecrets) Count() int {
	return len(k.values)
}

// SecretForms returns the raw, URL-escaped, JSON-escaped and base64 forms of a value, deduplicated.
func SecretForms(value string) []string {
	if value == "" {
		return []string{}
	}
	forms := []string{
		value,
		url.QueryEscape(value),
		url.PathEscape(value),
		strings.ReplaceAll(url.QueryEscape(value), "+", "%20"),
		jsonEscaped(value, true),
		jsonEscaped(value, false),
		base64.StdEncoding.EncodeToString([]byte(value)),
	}
	slices.Sort(forms)
	return slices.Compact(forms)
}

func (k KnownSecrets) Forms() []string {
	forms := make([]string, 0, len(k.values)*4)
	for value := range k.values {
		forms = append(forms, SecretForms(value)...)
	}
	slices.Sort(forms)
	forms = slices.Compact(forms)
	slices.SortStableFunc(forms, func(a, b string) int {
		return cmp.Compare(len(b), len(a))
	})
	return forms
}

// Mask hides every known form longest-first; forms under 4 characters only as a whole JSON string value.
func (k KnownSecrets) Mask(text string) string {
	masked := text
	for _, form := range k.Forms() {
		if len([]rune(form)) >= MinSecretMaskLength {
			masked = strings.ReplaceAll(masked, form, MaskedValue)
			continue
		}
		masked = strings.ReplaceAll(masked, jsonQuote+form+jsonQuote, jsonQuote+MaskedValue+jsonQuote)
	}
	return masked
}

func (k KnownSecrets) LogValue() any {
	return map[string]any{"count": len(k.values)}
}

func jsonEscaped(value string, escapeHTML bool) string {
	var quoted bytes.Buffer
	encoder := json.NewEncoder(&quoted)
	encoder.SetEscapeHTML(escapeHTML)
	_ = encoder.Encode(value)
	return strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(quoted.String()), jsonQuote), jsonQuote)
}
