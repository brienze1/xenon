package errors

import (
	"errors"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestAICallFailuresCarryTheirCodes(t *testing.T) {
	cases := map[string]error{
		"MCP-07400": RejectUnknownAIAction("tree", "explode", []string{"list"}),
		"MCP-08404": RejectMissingAIItem("01X"),
		"MCP-03403": RejectDeniedAISend(),
		"MCP-04408": RejectTimedOutAISend(),
		"MCP-06404": RejectInactiveAIWorkspace(),
		"MCP-02403": RejectHumanOnlyAIAction("Deleting a workspace"),
		"UNK-02500": FailUnknownAICall(errors.New("raw")),
	}
	for code, err := range cases {
		var custom errs.BaseError
		if !errors.As(err, &custom) || custom.Code() != code {
			t.Errorf("%s: got %v", code, err)
		}
	}
	for _, err := range []error{RejectInvalidArguments("tree", "bad"), RejectUnsupportedAIAction("tree", "list")} {
		var custom errs.BaseError
		if !errors.As(err, &custom) || custom.Code() != entity.AIInvalidArgumentsCode {
			t.Errorf("invalid arguments: got %v", err)
		}
	}
	if entity.AIActivityOutcomeOf(RejectTimedOutAISend().(errs.BaseError).Code()).String() != "timed_out" {
		t.Error("a timed-out send must record the timed_out outcome")
	}
}
