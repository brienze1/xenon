package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type RequestEditController interface {
	ApplyUrl(ctx context.Context, request dto.ApplyRequest) (*dto.Request, error)
	ApplyParams(ctx context.Context, request dto.ApplyRequest) (*dto.Request, error)
	ParseBulk(ctx context.Context, request dto.BulkText) (*dto.BulkRows, error)
	ParseCurl(ctx context.Context, request dto.CurlText) (*dto.ParsedCurl, error)
	FormatBulk(ctx context.Context, request dto.BulkRows) (*dto.BulkText, error)
	AutoHeaders(ctx context.Context, request dto.AutoHeadersRequest) (dto.AutoHeaderList, error)
}
