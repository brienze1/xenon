package usecase

import "context"

type FindActiveRunControls interface {
	FindControlById(ctx context.Context, localId, runId string) (*ActiveRun, error)
	FindControlsByWorkspace(ctx context.Context, localId string) []ActiveRun
	FindAllControls(ctx context.Context) []ActiveRun
}
