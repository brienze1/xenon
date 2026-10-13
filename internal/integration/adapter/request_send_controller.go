package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type RequestSendController interface {
	Send(ctx context.Context, request dto.SendRequest) (*dto.SendResult, error)
}
