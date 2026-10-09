package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var fileWatcherOnce sync.Once
var fileWatcherInstance adapter.FileWatcher

func GetFileWatcher() adapter.FileWatcher {
	if fileWatcherInstance == nil {
		fileWatcherOnce.Do(func() {
			fileWatcherInstance = infraAdapter.NewFileWatcher()
		})
	}
	return fileWatcherInstance
}

func SetFileWatcher(watcher adapter.FileWatcher) {
	fileWatcherOnce.Do(func() {})
	fileWatcherInstance = watcher
}
