package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/external"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var keychainWebserviceOnce sync.Once
var keychainWebserviceInstance adapter.KeychainWebservice

func GetKeychainWebservice() adapter.KeychainWebservice {
	if keychainWebserviceInstance == nil {
		keychainWebserviceOnce.Do(func() {
			keychainWebserviceInstance = webservice.NewKeychainWebservice(external.GetKeychain())
		})
	}
	return keychainWebserviceInstance
}
