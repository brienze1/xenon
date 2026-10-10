package agecrypt

import (
	"bytes"
	"io"

	"filippo.io/age"
)

const maxScryptWorkFactor = 22

// GenerateIdentity returns a new X25519 identity in its "AGE-SECRET-KEY-1…" text form.
func GenerateIdentity() (string, error) {
	identity, err := age.GenerateX25519Identity()
	if err != nil {
		return "", err
	}
	return identity.String(), nil
}

// EncryptForIdentity encrypts plaintext so that only the given identity can decrypt it.
func EncryptForIdentity(plaintext []byte, identity string) ([]byte, error) {
	parsed, err := age.ParseX25519Identity(identity)
	if err != nil {
		return nil, err
	}
	return encrypt(plaintext, parsed.Recipient())
}

func DecryptWithIdentity(ciphertext []byte, identity string) ([]byte, error) {
	parsed, err := age.ParseX25519Identity(identity)
	if err != nil {
		return nil, err
	}
	return decrypt(ciphertext, parsed)
}

// EncryptWithPassphrase derives the key with scrypt at work factor 2^logN.
func EncryptWithPassphrase(plaintext []byte, passphrase string, logN int) ([]byte, error) {
	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return nil, err
	}
	recipient.SetWorkFactor(logN)
	return encrypt(plaintext, recipient)
}

func DecryptWithPassphrase(ciphertext []byte, passphrase string) ([]byte, error) {
	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return nil, err
	}
	identity.SetMaxWorkFactor(maxScryptWorkFactor)
	return decrypt(ciphertext, identity)
}

func encrypt(plaintext []byte, recipient age.Recipient) ([]byte, error) {
	var output bytes.Buffer
	writer, err := age.Encrypt(&output, recipient)
	if err != nil {
		return nil, err
	}
	if _, err := writer.Write(plaintext); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func decrypt(ciphertext []byte, identity age.Identity) ([]byte, error) {
	reader, err := age.Decrypt(bytes.NewReader(ciphertext), identity)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(reader)
}
