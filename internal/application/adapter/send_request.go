package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SendRequestService interface {
	Send(ctx context.Context, localId, tabId string, request entity.Request) (*entity.Response, error)
}
