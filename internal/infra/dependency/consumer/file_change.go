package consumer

import (
	"context"
	"sync"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/infra/dependency/integration"
	"github.com/brienze1/xenon/internal/infra/server/consumer"
)

var fileChangeConsumerOnce sync.Once
var fileChangeConsumerInstance *consumer.FileChangeConsumer

func GetFileChangeConsumer() *consumer.FileChangeConsumer {
	if fileChangeConsumerInstance == nil {
		fileChangeConsumerOnce.Do(func() {
			fileChangeConsumerInstance = consumer.NewFileChangeConsumer(controller.GetFileChangeController(), integration.GetWorkspaceWatcherConsumerWebservice())
			fileChangeConsumerInstance.Register(context.Background())
		})
	}
	return fileChangeConsumerInstance
}
