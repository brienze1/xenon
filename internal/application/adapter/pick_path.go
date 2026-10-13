package adapter

import "context"

type PickPathService interface {
	PickFolder(ctx context.Context, title string) (string, error)
	PickFile(ctx context.Context, title string) (string, error)
	PickSaveFile(ctx context.Context, title, suggestedName string) (string, error)
}
