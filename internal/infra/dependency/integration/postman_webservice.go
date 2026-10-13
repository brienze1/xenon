package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice"
)

var postmanWebserviceOnce sync.Once
var postmanWebserviceInstance adapter.PostmanWebservice

func GetPostmanWebservice() adapter.PostmanWebservice {
	if postmanWebserviceInstance == nil {
		postmanWebserviceOnce.Do(func() {
			postmanWebserviceInstance = webservice.NewPostmanWebservice()
		})
	}
	return postmanWebserviceInstance
}
