package timeutils

import (
	"sync"
	"time"
)

type CurrentTimeProvider func() time.Time

type timeConfig struct {
	currentTimeProvider CurrentTimeProvider
	mu                  sync.RWMutex
}

var instance *timeConfig
var once sync.Once

func getTimeConfig() *timeConfig {
	once.Do(func() {
		instance = &timeConfig{currentTimeProvider: time.Now}
	})
	return instance
}

func SetCurrentTimeProvider(provider CurrentTimeProvider) {
	config := getTimeConfig()
	config.mu.Lock()
	defer config.mu.Unlock()
	config.currentTimeProvider = provider
}

func (t *timeConfig) provider() CurrentTimeProvider {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentTimeProvider
}
