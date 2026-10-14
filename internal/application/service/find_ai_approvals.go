package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type findAIApprovals struct {
	aiApprovals usecase.FindAIApprovals
}

func NewFindAIApprovalsService(aiApprovals usecase.FindAIApprovals) adapter.FindAIApprovalsService {
	return &findAIApprovals{aiApprovals: aiApprovals}
}

func (f *findAIApprovals) FindPending(ctx context.Context) []entity.AIApproval {
	logger.Info(ctx, "Started")

	pending := f.aiApprovals.FindPending(ctx)

	logger.Info(ctx, "Finished", pending)
	return pending
}
