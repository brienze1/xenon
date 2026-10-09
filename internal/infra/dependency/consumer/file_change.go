package consumer

import (
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
	"github.com/brienze1/xenon/internal/infra/server/adapter"
	"github.com/brienze1/xenon/internal/infra/server/consumer"
)

var fileChangeConsumerOnce sync.Once
var fileChangeConsumerInstance adapter.Consumer

func GetFileChangeConsumer() adapter.Consumer {
	if fileChangeConsumerInstance == nil {
		fileChangeConsumerOnce.Do(func() {
			fileChangeConsumerInstance = consumer.NewFileChangeConsumer(controller.GetFileChangeController(), integration.GetWorkspaceWatcherConsumerWebservice())
		})
	}
	return fileChangeConsumerInstance
}
