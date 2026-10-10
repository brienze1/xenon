package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var archiveOnce sync.Once
var archiveInstance adapter.Archive

func GetArchive() adapter.Archive {
	if archiveInstance == nil {
		archiveOnce.Do(func() {
			archiveInstance = infraAdapter.NewArchive()
		})
	}
	return archiveInstance
}
