package usecase

import "context"

type GenerateDynamicValue interface {
	Generate(ctx context.Context, name string) (string, bool)
	Names(ctx context.Context) []string
}
