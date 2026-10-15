package usecase

import "context"

type RestartApp interface {
	RestartApp(ctx context.Context, executable string) error
}
