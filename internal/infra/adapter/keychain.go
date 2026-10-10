package adapter

import (
	"errors"

	"github.com/zalando/go-keyring"

	"github.com/brienze1/xenon/internal/integration/adapter"
)

type osKeychain struct{}

func NewKeychain() adapter.Keychain {
	return &osKeychain{}
}

func (o *osKeychain) Get(service, account string) (string, bool, error) {
	secret, err := keyring.Get(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return secret, true, nil
}

func (o *osKeychain) Set(service, account, secret string) error {
	return keyring.Set(service, account, secret)
}

func (o *osKeychain) Delete(service, account string) error {
	if err := keyring.Delete(service, account); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}
	return nil
}
