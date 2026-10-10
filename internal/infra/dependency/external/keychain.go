package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var keychainOnce sync.Once
var keychainInstance adapter.Keychain

func GetKeychain() adapter.Keychain {
	if keychainInstance == nil {
		keychainOnce.Do(func() {
			keychainInstance = infraAdapter.NewKeychain()
		})
	}
	return keychainInstance
}

func SetKeychain(keychain adapter.Keychain) {
	keychainOnce.Do(func() {})
	keychainInstance = keychain
}
