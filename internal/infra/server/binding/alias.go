package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type AliasService struct {
	manage adapter.AliasManageController
}

func NewAliasService(manage adapter.AliasManageController) *AliasService {
	return &AliasService{manage: manage}
}

func (a *AliasService) Create(ctx context.Context, request dto.CreateAlias) (*dto.Request, error) {
	return a.manage.Create(ctx, request)
}

func (a *AliasService) Save(ctx context.Context, request dto.SaveAlias) (*dto.Request, error) {
	return a.manage.Save(ctx, request)
}

func (a *AliasService) Delete(ctx context.Context, request dto.AliasReference) (*dto.Request, error) {
	return a.manage.Delete(ctx, request)
}
