package usecase

import "context"

type ListCommitMessages interface {
	ListMessages(ctx context.Context, clonePath, base, head string) ([]string, error)
}
