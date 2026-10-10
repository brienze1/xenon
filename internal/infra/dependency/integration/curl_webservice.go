package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/curl"
)

var curlWebserviceOnce sync.Once
var curlWebserviceInstance adapter.CurlWebservice

func GetCurlWebservice() adapter.CurlWebservice {
	if curlWebserviceInstance == nil {
		curlWebserviceOnce.Do(func() {
			curlWebserviceInstance = curl.NewCurlWebservice()
		})
	}
	return curlWebserviceInstance
}
