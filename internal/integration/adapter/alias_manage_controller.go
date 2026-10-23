package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type AliasManageController interface {
	Create(ctx context.Context, request dto.CreateAlias) (*dto.Request, error)
	Save(ctx context.Context, request dto.SaveAlias) (*dto.Request, error)
	SetEnvironments(ctx context.Context, request dto.SetAliasEnvironments) (*dto.Request, error)
	Delete(ctx context.Context, request dto.AliasReference) (*dto.Request, error)
}
