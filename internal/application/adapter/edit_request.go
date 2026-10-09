package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type EditRequestService interface {
	ApplyUrl(ctx context.Context, request entity.Request) entity.Request
	ApplyParams(ctx context.Context, request entity.Request) entity.Request
	ParseBulk(ctx context.Context, text string) []entity.KeyValue
	FormatBulk(ctx context.Context, rows []entity.KeyValue) string
	AutoHeaders(ctx context.Context, localId string, request entity.Request, appVersion string) ([]entity.AutoHeader, error)
}
