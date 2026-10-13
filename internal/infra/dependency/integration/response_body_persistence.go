package integration

import (
	"sync"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/persistence"
)

var responseBodyPersistenceOnce sync.Once
var responseBodyPersistenceInstance adapter.ResponseBodyPersistence

func GetResponseBodyPersistence() adapter.ResponseBodyPersistence {
	if responseBodyPersistenceInstance == nil {
		responseBodyPersistenceOnce.Do(func() {
			responseBodyPersistenceInstance = persistence.NewResponseBodyPersistence()
		})
	}
	return responseBodyPersistenceInstance
}
