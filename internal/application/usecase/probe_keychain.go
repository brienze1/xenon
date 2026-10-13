package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ProbeKeychain interface {
	Probe(ctx context.Context) entity.KeychainProbe
}
