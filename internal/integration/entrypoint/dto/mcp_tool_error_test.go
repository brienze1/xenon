package dto

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

func TestMCPToolErrorNamesTheArgumentTheClientSent(t *testing.T) {
	tool := testTool("workspace_edit", []string{"create", "open", "switch", "rename"}, NewMCPToolArguments[testEditArguments]("create", "open", "switch", "rename"))
	failure := errs.BadRequestError("invalid request: local_id is required", "WKS-08400", errs.ErrorDetails{Attribute: "local_id", Messages: []string{"is required"}})
	got := NewMCPToolError(failure, tool, "switch")
	if got.Code != "WKS-08400" || got.Message != "invalid request: id is required" {
		t.Errorf("NewMCPToolError() = %+v", got)
	}
	if len(got.Details) != 1 || got.Details[0] != (MCPToolErrorDetail{Field: "id", Problem: "is required"}) {
		t.Errorf("Details = %+v", got.Details)
	}
}

func TestMCPToolErrorPassesOtherDetailsThrough(t *testing.T) {
	failure := errs.ConflictError("some secrets are missing", "SND-06409", errs.ErrorDetails{Attribute: "missing_secrets", Messages: []string{"A", "B"}})
	got := NewMCPToolError(failure, MCPTool{}, "")
	if got.Message != "some secrets are missing" {
		t.Errorf("Message = %q, want the original message", got.Message)
	}
	if len(got.Details) != 1 || got.Details[0] != (MCPToolErrorDetail{Field: "missing_secrets", Problem: "A, B"}) {
		t.Errorf("Details = %+v", got.Details)
	}
}

func TestMCPToolErrorWithoutDetailsOmitsTheList(t *testing.T) {
	got := NewMCPToolError(errs.NotFoundError("No item with id \"x\" in the active workspace", "MCP-08404"), MCPTool{}, "")
	encoded, _ := json.Marshal(got)
	if strings.Contains(string(encoded), "details") {
		t.Errorf("encoded = %s, want no details key", encoded)
	}
}
