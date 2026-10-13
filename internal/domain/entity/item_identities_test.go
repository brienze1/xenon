package entity

import (
	"strconv"
	"testing"
)

func TestIdentifyKeepsFirstIdAndReidentifiesDuplicatesStably(t *testing.T) {
	generated := 0
	newId := func() string {
		generated++
		return "new-" + strconv.Itoa(generated)
	}
	previous := map[string]string{"/b": "kept"}
	identities := NewItemIdentities(previous)
	if id, warning := identities.Identify("same", "/a", newId); id != "same" || warning != "" {
		t.Fatalf("first = %q %q", id, warning)
	}
	if id, warning := identities.Identify("same", "/b", newId); id != "kept" || warning != duplicateIdWarning {
		t.Fatalf("duplicate = %q %q", id, warning)
	}
	id, warning := identities.Identify("same", "/c", newId)
	if id != "new-1" || warning != duplicateIdWarning || identities.Reidentified()["/c"] != id {
		t.Fatalf("new duplicate = %q %q", id, warning)
	}
	missing, warning := identities.Identify("", "/d", newId)
	if missing != "new-2" || warning != missingIdWarning {
		t.Fatalf("missing id = %q %q", missing, warning)
	}
	previous["/e"] = "kept"
	if id, _ := identities.Identify("", "/e", newId); id == "kept" {
		t.Fatal("a previous id already in use must not be reused")
	}
}
