package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/pkg/logger"
)

type publishAIAccessStatus struct {
	findAIAccessStatus             usecase.FindAIAccessStatus
	replacePublishedAIAccessStatus usecase.ReplacePublishedAIAccessStatus
	publishAIEvent                 usecase.PublishAIEvent
}

func NewPublishAIAccessStatusService(
	findAIAccessStatus usecase.FindAIAccessStatus,
	replacePublishedAIAccessStatus usecase.ReplacePublishedAIAccessStatus,
	publishAIEvent usecase.PublishAIEvent,
) adapter.PublishAIAccessStatusService {
	return &publishAIAccessStatus{
		findAIAccessStatus:             findAIAccessStatus,
		replacePublishedAIAccessStatus: replacePublishedAIAccessStatus,
		publishAIEvent:                 publishAIEvent,
	}
}

func (p *publishAIAccessStatus) Publish(ctx context.Context) error {
	logger.Info(ctx, "Started")

	status := p.findAIAccessStatus.Status(ctx)
	if status.Differs(p.replacePublishedAIAccessStatus.ReplacePublished(ctx, status)) {
		if err := p.publishAIEvent.AccessStatusChanged(ctx, status); err != nil {
			return err
		}
	}

	logger.Info(ctx, "Finished", status)
	return nil
}
