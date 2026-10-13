package usecase

import "context"

type FocusWindow interface {
	Focus(ctx context.Context)
}
