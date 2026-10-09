package adapter

import "context"

type FileWatcher interface {
	Watch(ctx context.Context, root string, onChange func(path string, removed bool)) error
	Stop() error
}
