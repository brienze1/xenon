package external

import (
	"sync"

	infraAdapter "github.com/brienze1/xenon/internal/infra/adapter"
	"github.com/brienze1/xenon/internal/integration/adapter"
)

var trashOnce sync.Once
var trashInstance adapter.Trash

func GetTrash() adapter.Trash {
	if trashInstance == nil {
		trashOnce.Do(func() {
			trashInstance = infraAdapter.NewTrash()
		})
	}
	return trashInstance
}

func SetTrash(trash adapter.Trash) {
	trashOnce.Do(func() {})
	trashInstance = trash
}
