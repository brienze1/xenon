package entity

import "testing"

func TestAIOutputHasSecretsComparesTheGuardedText(t *testing.T) {
	if AIOutputHasSecrets("plain", "plain") {
		t.Error("an unchanged text has no secrets")
	}
	if !AIOutputHasSecrets("token=abc", "token=•••••") {
		t.Error("a masked text has secrets")
	}
}
