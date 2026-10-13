package adapter

import "context"

type HandleFileChangeService interface {
	Handle(ctx context.Context, localId, path string) error
}
