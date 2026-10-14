package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type findAIAccessStatus struct {
	aiAccessStatus usecase.FindAIAccessStatus
}

func NewFindAIAccessStatusService(aiAccessStatus usecase.FindAIAccessStatus) adapter.FindAIAccessStatusService {
	return &findAIAccessStatus{aiAccessStatus: aiAccessStatus}
}

func (f *findAIAccessStatus) Find(ctx context.Context) entity.AIAccessStatus {
	logger.Info(ctx, "Started")

	status := f.aiAccessStatus.Status(ctx)

	logger.Info(ctx, "Finished", status)
	return status
}
