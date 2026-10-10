package mock

import (
	"errors"
	"sync"
)

type keychainMode int

const (
	keychainAvailable keychainMode = iota
	keychainUnavailable
	keychainFailing
)

var errKeychainUnsupported = errors.New("keychain: unsupported platform")

type Keychain struct {
	entries map[string]string
	failure error
	mode    keychainMode
	mu      sync.RWMutex
}

func NewKeychain() *Keychain {
	return &Keychain{entries: map[string]string{}}
}

func (k *Keychain) Get(service, account string) (string, bool, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	if err := k.unusable(); err != nil {
		return "", false, err
	}
	secret, found := k.entries[entryName(service, account)]
	return secret, found, nil
}

func (k *Keychain) Set(service, account, secret string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if err := k.unusable(); err != nil {
		return err
	}
	k.entries[entryName(service, account)] = secret
	return nil
}

func (k *Keychain) Delete(service, account string) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	if err := k.unusable(); err != nil {
		return err
	}
	delete(k.entries, entryName(service, account))
	return nil
}

func (k *Keychain) MakeAvailable() {
	k.setMode(keychainAvailable, nil)
}

func (k *Keychain) MakeUnavailable() {
	k.setMode(keychainUnavailable, errKeychainUnsupported)
}

func (k *Keychain) Fail(message string) {
	k.setMode(keychainFailing, errors.New(message))
}

func (k *Keychain) HasAccount(account string) bool {
	k.mu.RLock()
	defer k.mu.RUnlock()
	for name := range k.entries {
		if accountOf(name) == account {
			return true
		}
	}
	return false
}

func (k *Keychain) RemoveAccount(account string) {
	k.mu.Lock()
	defer k.mu.Unlock()
	for name := range k.entries {
		if accountOf(name) == account {
			delete(k.entries, name)
		}
	}
}

func (k *Keychain) Count() int {
	k.mu.RLock()
	defer k.mu.RUnlock()
	return len(k.entries)
}

func (k *Keychain) Reset() {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.entries = map[string]string{}
	k.mode = keychainAvailable
	k.failure = nil
}

func (k *Keychain) setMode(mode keychainMode, failure error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.mode = mode
	k.failure = failure
}

func (k *Keychain) unusable() error {
	if k.mode == keychainAvailable {
		return nil
	}
	return k.failure
}

func entryName(service, account string) string {
	return service + "\x00" + account
}

func accountOf(name string) string {
	for index := 0; index < len(name); index++ {
		if name[index] == 0 {
			return name[index+1:]
		}
	}
	return name
}
