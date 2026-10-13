package enums_test

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/aiaccessstate"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/approvaloutcome"
	"github.com/brienze1/xenon/internal/domain/enums/historyfilter"
	"github.com/brienze1/xenon/internal/domain/enums/historykind"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
)

func TestMcpEnumsParseTheirValues(t *testing.T) {
	cases := []struct {
		got  string
		want string
	}{
		{historysource.GetHistorySource(" AI ").String(), "ai"},
		{historysource.GetHistorySource("user").String(), "user"},
		{historysource.GetHistorySource("").String(), "user"},
		{historykind.GetHistoryKind("action").String(), "action"},
		{historykind.GetHistoryKind("").String(), "send"},
		{historyfilter.GetHistoryFilter("USER").String(), "user"},
		{historyfilter.GetHistoryFilter("ai").String(), "ai"},
		{historyfilter.GetHistoryFilter("other").String(), "all"},
		{aiaccessstate.GetAIAccessState("running").String(), "running"},
		{aiaccessstate.GetAIAccessState("error").String(), "error"},
		{aiaccessstate.GetAIAccessState("").String(), "off"},
		{approvaloutcome.GetApprovalOutcome("approved").String(), "approved"},
		{approvaloutcome.GetApprovalOutcome("denied").String(), "denied"},
		{approvaloutcome.GetApprovalOutcome("timed_out").String(), "timed_out"},
		{approvaloutcome.GetApprovalOutcome("").String(), "pending"},
		{aichangekind.GetAIChangeKind("tree").String(), "tree"},
		{aichangekind.GetAIChangeKind("request").String(), "request"},
		{aichangekind.GetAIChangeKind("variables").String(), "variables"},
		{aichangekind.GetAIChangeKind("history").String(), "history"},
		{aichangekind.GetAIChangeKind("cookies").String(), "cookies"},
		{aichangekind.GetAIChangeKind("workspaces").String(), "workspaces"},
		{aichangekind.GetAIChangeKind("git").String(), "unknown"},
	}
	for _, test := range cases {
		if test.got != test.want {
			t.Errorf("got %q, want %q", test.got, test.want)
		}
	}
	if !historysource.AI.Equals("AI") || historykind.Send.Equals("action") || !aiaccessstate.Off.Equals("off") {
		t.Error("Equals must compare case-insensitively")
	}
}
