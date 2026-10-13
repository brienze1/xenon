package external

import (
	"sync"

	httpClient "github.com/brienze1/xenon/pkg/http"
	"github.com/brienze1/xenon/pkg/logger"
)

var httpClientFactoryOnce sync.Once
var httpClientFactoryInstance httpClient.ClientFactory

func GetHttpClientFactory() httpClient.ClientFactory {
	if httpClientFactoryInstance == nil {
		httpClientFactoryOnce.Do(func() {
			httpClientFactoryInstance = httpClient.HttpClientFactory(logger.LoggerWrapper())
		})
	}
	return httpClientFactoryInstance
}

func SetHttpClientFactory(factory httpClient.ClientFactory) {
	httpClientFactoryOnce.Do(func() {})
	httpClientFactoryInstance = factory
}
