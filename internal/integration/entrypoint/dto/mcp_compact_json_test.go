package dto

import "testing"

func TestEncodeMCPTextOmitsEmptyFieldsButKeepsItems(t *testing.T) {
	type row struct {
		Name   string   `json:"name"`
		Note   string   `json:"note"`
		Passed bool     `json:"passed"`
		Tags   []string `json:"tags"`
	}
	payload := map[string]any{"items": []row{}, "extra": map[string]any{}, "html": "<a&b>", "rows": []row{{Name: "x"}}, "pair": []string{"k", ""}, "nothing": nil}
	encoded, err := EncodeMCPText(payload)
	if err != nil || encoded != `{"html":"<a&b>","items":[],"pair":["k",""],"rows":[{"name":"x","passed":false}]}` {
		t.Fatalf("EncodeMCPText() = %s, %v", encoded, err)
	}
	if empty, _ := EncodeMCPText(map[string]any{}); empty != `{}` {
		t.Errorf("an empty result = %s", empty)
	}
	if _, err := EncodeMCPText(func() {}); err == nil {
		t.Error("an unencodable payload must fail")
	}
}
