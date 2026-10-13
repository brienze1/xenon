package formdatatype

import "testing"

func TestGetFormDataType(t *testing.T) {
	cases := map[string]any{"file": File, " FILE ": File, "text": Text, " Text ": Text, "": Unknown, "other": Unknown}
	for raw, want := range cases {
		if got := GetFormDataType(raw); got != want {
			t.Errorf("GetFormDataType(%q) = %q, want %q", raw, got, want)
		}
	}
	if !File.Equals("File") || File.Equals("text") || File.String() != "file" {
		t.Error("Equals/String misbehave")
	}
}
