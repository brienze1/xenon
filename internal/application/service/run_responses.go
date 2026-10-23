package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
)

func releaseRunResponses(ctx context.Context, releaseResponse usecase.ReleaseResponse, runs ...entity.Run) error {
	for _, run := range runs {
		for _, key := range run.ResponseKeys() {
			if err := releaseResponse.Release(ctx, key); err != nil {
				return err
			}
		}
	}
	return nil
}
