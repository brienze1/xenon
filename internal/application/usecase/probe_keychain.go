package usecase

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type ProbeKeychain interface {
	Probe(ctx context.Context) enums.VaultMode
}
