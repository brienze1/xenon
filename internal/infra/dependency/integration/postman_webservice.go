package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/postman"
)

var postmanWebserviceOnce sync.Once
var postmanWebserviceInstance adapter.PostmanWebservice

func GetPostmanWebservice() adapter.PostmanWebservice {
	if postmanWebserviceInstance == nil {
		postmanWebserviceOnce.Do(func() {
			postmanWebserviceInstance = postman.NewPostmanWebservice()
		})
	}
	return postmanWebserviceInstance
}
