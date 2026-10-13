package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var awsV4WebserviceOnce sync.Once
var awsV4WebserviceInstance adapter.AwsV4Webservice

func GetAwsV4Webservice() adapter.AwsV4Webservice {
	if awsV4WebserviceInstance == nil {
		awsV4WebserviceOnce.Do(func() {
			awsV4WebserviceInstance = webservice.NewAwsV4Webservice()
		})
	}
	return awsV4WebserviceInstance
}
