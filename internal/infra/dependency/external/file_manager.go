package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var fileManagerOnce sync.Once
var fileManagerInstance adapter.FileManager

func GetFileManager() adapter.FileManager {
	if fileManagerInstance == nil {
		fileManagerOnce.Do(func() {
			fileManagerInstance = infraAdapter.NewFileManager()
		})
	}
	return fileManagerInstance
}

func SetFileManager(fileManager adapter.FileManager) {
	fileManagerOnce.Do(func() {})
	fileManagerInstance = fileManager
}
