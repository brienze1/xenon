package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

const (
	mcpSendTargetConflictReason = "pass request_id or request, not both"
	mcpSendTargetMissingReason  = "request_id or request is required"
	mcpSendTargetAliasReason    = "alias_id needs request_id"
)

func resolveMCPSendTarget(
	ctx context.Context,
	searchTreeService appAdapter.SearchTreeService,
	saveRequestService appAdapter.SaveRequestService,
	call dto.MCPToolCall,
	target dto.MCPSendTarget,
) (dto.Request, entity.AIActivityTarget, error) {
	if target.IsAmbiguous() {
		return dto.Request{}, entity.AIActivityTarget{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpSendTargetConflictReason)
	}
	if target.HasStrayAlias() {
		return dto.Request{}, entity.AIActivityTarget{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpSendTargetAliasReason)
	}
	if target.IsAdHoc() {
		return target.AdHocRequest(), target.AdHocTarget(), nil
	}
	if target.IsMissing() {
		return dto.Request{}, entity.AIActivityTarget{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpSendTargetMissingReason)
	}

	requestId := target.RequestItemId()
	tree, err := loadMCPTree(ctx, searchTreeService, call.LocalId)
	if err != nil {
		return dto.Request{}, entity.AIActivityTarget{}, err
	}
	entry, found := tree.Entry(requestId)
	if !found || !entry.IsRequest() {
		return dto.Request{}, entity.NewAIItemTarget(requestId), mcpErrors.RejectMissingAIItem(requestId)
	}

	request := dto.NewOpenRequest(call.LocalId, requestId)
	err = validateOpenRequest(&request)
	if err != nil {
		return dto.Request{}, entry.Target(), err
	}

	saved, err := saveRequestService.Open(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.Request{}, entry.Target(), err
	}
	return dto.NewRequest(saved.Version(target.AliasId)), entry.Target(), nil
}
