package usecase

import "context"

type CopyWorkspaceFolder interface {
	CopyFolder(ctx context.Context, source, target string) error
	CopyInto(ctx context.Context, source, target string) error
}
