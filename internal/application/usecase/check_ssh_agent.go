package usecase

import "context"

type CheckSshAgent interface {
	CheckAgent(ctx context.Context) error
}
