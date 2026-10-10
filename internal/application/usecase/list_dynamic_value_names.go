package usecase

import "context"

type ListDynamicValueNames interface {
	Names(ctx context.Context) []string
}
