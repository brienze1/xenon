package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestAIApprovalPersistenceRefusesUnknownAndResolvedApprovals(t *testing.T) {
	ctx := context.Background()
	approvals := NewAIApprovalPersistence()
	approval := entity.NewAIApproval("01A", "Claude Code", "GET", "https://example.test", entity.AIApprovalEnvironment{}, time.Now())
	resolved := approvals.Queue(ctx, approval, "s1")

	if answered, err := approvals.Resolve(ctx, "01A", entity.AIApproval.Denied); err != nil || (<-resolved).Outcome != answered.Outcome {
		t.Fatalf("first resolve = %+v, %v", answered, err)
	}
	cases := map[string]string{"01A": "MCP-05409", "unknown": "MCP-07400"}
	for id, code := range cases {
		var custom errs.BaseError
		if answered, err := approvals.Resolve(ctx, id, entity.AIApproval.Denied); answered != nil || !errors.As(err, &custom) || custom.Code() != code {
			t.Errorf("%s: got %+v, %v", id, answered, err)
		}
	}
}

func TestAIApprovalPersistenceSettlesOnlyPendingApprovals(t *testing.T) {
	ctx := context.Background()
	approvals := NewAIApprovalPersistence()
	approval := entity.NewAIApproval("01A", "Claude Code", "GET", "https://example.test", entity.AIApprovalEnvironment{}, time.Now())
	resolved := approvals.Queue(ctx, approval, "s1")

	if settled, ok := approvals.Settle(ctx, "01A", entity.AIApproval.TimedOut); !ok || (<-resolved).Outcome != settled.Outcome || len(approvals.FindPending(ctx)) != 0 {
		t.Fatalf("first settle = %+v, %v", settled, ok)
	}
	for _, id := range []string{"01A", "unknown"} {
		if settled, ok := approvals.Settle(ctx, id, entity.AIApproval.Denied); settled != nil || ok {
			t.Errorf("%s: got %+v, %v", id, settled, ok)
		}
	}
}
