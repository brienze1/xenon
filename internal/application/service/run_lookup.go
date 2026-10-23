package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type runLookup struct {
	findActiveRuns usecase.FindActiveRuns
	findRuns       usecase.FindRuns
	findResponse   usecase.FindResponse
}

func (l runLookup) held(ctx context.Context, localId, runId string) (*entity.Run, error) {
	run, err := l.run(ctx, localId, runId)
	if err != nil {
		return nil, err
	}
	released, err := l.releasedResponses(ctx, *run)
	if err != nil {
		return nil, err
	}
	held := run.WithoutResponses(released)
	return &held, nil
}

func (l runLookup) run(ctx context.Context, localId, runId string) (*entity.Run, error) {
	if active, found := l.findActiveRuns.FindById(ctx, localId, runId); found {
		return active, nil
	}
	return l.findRuns.FindRun(ctx, localId, runId)
}

func (l runLookup) releasedResponses(ctx context.Context, run entity.Run) ([]string, error) {
	released := make([]string, 0)
	for _, key := range run.ResponseKeys() {
		_, err := l.findResponse.Find(ctx, key)
		if errs.IsNotFoundError(err) {
			released = append(released, key)
			continue
		}
		if err != nil {
			return nil, err
		}
	}
	return released, nil
}
