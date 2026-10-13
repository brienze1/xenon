package usecase

import "context"

type CopyToClipboard interface {
	Copy(ctx context.Context, text string) error
}
