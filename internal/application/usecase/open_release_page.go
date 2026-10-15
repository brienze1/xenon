package usecase

import "context"

type OpenReleasePage interface {
	OpenReleasePage(ctx context.Context, url string) error
}
