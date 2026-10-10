package encryption

import (
	"bytes"
	"strings"
	"testing"
)

func TestIdentityRoundTrip(t *testing.T) {
	identity, err := GenerateIdentity()
	if err != nil || !strings.HasPrefix(identity, "AGE-SECRET-KEY-1") {
		t.Fatalf("GenerateIdentity() = %q, %v", identity, err)
	}
	ciphertext, err := EncryptForIdentity([]byte("s3cr3t"), identity)
	if err != nil || bytes.Contains(ciphertext, []byte("s3cr3t")) {
		t.Fatalf("EncryptForIdentity() leaked or failed: %v", err)
	}
	plaintext, err := DecryptWithIdentity(ciphertext, identity)
	if err != nil || string(plaintext) != "s3cr3t" {
		t.Fatalf("DecryptWithIdentity() = %q, %v", plaintext, err)
	}
	other, _ := GenerateIdentity()
	if _, err := DecryptWithIdentity(ciphertext, other); err == nil {
		t.Fatal("another identity must not decrypt")
	}
	if _, err := DecryptWithIdentity([]byte("not age"), identity); err == nil {
		t.Fatal("garbage must not decrypt")
	}
}

func TestInvalidIdentities(t *testing.T) {
	if _, err := EncryptForIdentity([]byte("x"), "nope"); err == nil {
		t.Fatal("EncryptForIdentity must reject an invalid identity")
	}
	if _, err := DecryptWithIdentity([]byte("x"), "nope"); err == nil {
		t.Fatal("DecryptWithIdentity must reject an invalid identity")
	}
}

func TestPassphraseRoundTrip(t *testing.T) {
	ciphertext, err := EncryptWithPassphrase([]byte("identity"), "correct horse", 10)
	if err != nil {
		t.Fatal(err)
	}
	plaintext, err := DecryptWithPassphrase(ciphertext, "correct horse")
	if err != nil || string(plaintext) != "identity" {
		t.Fatalf("DecryptWithPassphrase() = %q, %v", plaintext, err)
	}
	if _, err := DecryptWithPassphrase(ciphertext, "wrong horse"); err == nil {
		t.Fatal("a wrong passphrase must not decrypt")
	}
	if _, err := EncryptWithPassphrase([]byte("x"), "", 10); err == nil {
		t.Fatal("an empty passphrase must be rejected")
	}
	if _, err := DecryptWithPassphrase(ciphertext, ""); err == nil {
		t.Fatal("an empty passphrase must be rejected")
	}
}
