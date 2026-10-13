package utils

import "testing"

func TestPtr(t *testing.T) {
	value := Ptr("xenon")
	if value == nil || *value != "xenon" {
		t.Fatalf("Ptr() = %v", value)
	}
	if Ptr(1) == Ptr(1) {
		t.Error("each call should return a new pointer")
	}
}
