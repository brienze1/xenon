package usecase

import "context"

type DeleteWorkspaceRuns interface {
	DeleteWorkspaceRuns(ctx context.Context, localId string) error
}
