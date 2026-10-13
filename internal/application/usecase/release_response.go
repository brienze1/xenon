package usecase

import "context"

type ReleaseResponse interface {
	Release(ctx context.Context, tabId string) error
	ReleaseAll(ctx context.Context) error
}
