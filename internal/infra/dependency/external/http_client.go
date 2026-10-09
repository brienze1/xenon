package external

import (
	"sync"

	httpClient "github.com/brienze1/xenon/pkg/http"
)

var httpClientFactoryOnce sync.Once
var httpClientFactoryInstance httpClient.ClientFactory

func GetHttpClientFactory() httpClient.ClientFactory {
	if httpClientFactoryInstance == nil {
		httpClientFactoryOnce.Do(func() {
			httpClientFactoryInstance = httpClient.NewClient
		})
	}
	return httpClientFactoryInstance
}
