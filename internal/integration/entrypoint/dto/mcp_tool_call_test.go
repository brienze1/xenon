package dto

import (
	"errors"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

func TestMCPToolCallDecodesItsArguments(t *testing.T) {
	var target struct {
		Name string `json:"name"`
	}
	call := NewMCPToolCall(NewMCPCall("s1", "client", []byte(`{"name":"x"}`)), "tree", "list", "ws-1")
	if err := call.Decode(&target); err != nil || target.Name != "x" {
		t.Errorf("Decode() = %v %+v", err, target)
	}
	if call.TabId() != "mcp:s1" {
		t.Errorf("TabId() = %s", call.TabId())
	}
	var custom errs.BaseError
	wrongType := NewMCPToolCall(NewMCPCall("s1", "", []byte(`{"name":1}`)), "tree", "list", "ws-1")
	if err := wrongType.Decode(&target); !errors.As(err, &custom) || custom.Code() != "MCP-07400" {
		t.Errorf("Decode() of a wrong type = %v", err)
	}
	if err := NewMCPToolCall(NewMCPCall("s1", "", nil), "tree", "list", "ws-1").Decode(&target); err != nil {
		t.Errorf("Decode() of no arguments = %v", err)
	}
}
