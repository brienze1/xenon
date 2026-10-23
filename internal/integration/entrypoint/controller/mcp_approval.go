package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

func askMCPApproval(ctx context.Context, approveAISendService appAdapter.ApproveAISendService, approvalRequest dto.AIApprovalRequest) error {
	err := validateAIArguments(&approvalRequest)
	if err != nil {
		return err
	}

	approval, err := approveAISendService.Request(ctx, approvalRequest.ToEntity(), approvalRequest.Session())
	if err != nil {
		return err
	}
	switch {
	case approval.IsApproved():
		return nil
	case approval.IsTimedOut():
		return mcpErrors.RejectTimedOutAISend()
	default:
		return mcpErrors.RejectDeniedAISend()
	}
}
