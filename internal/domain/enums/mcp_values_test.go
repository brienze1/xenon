package enums_test

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/aiaccessstate"
	"github.com/brienze1/xenon/internal/domain/enums/aiactivityoutcome"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
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
		{historysource.GetHistorySource("other").String(), "unknown"},
		{historykind.GetHistoryKind("action").String(), "action"},
		{historykind.GetHistoryKind("send").String(), "send"},
		{historykind.GetHistoryKind("").String(), "send"},
		{historykind.GetHistoryKind("other").String(), "unknown"},
		{historyfilter.GetHistoryFilter("USER").String(), "user"},
		{historyfilter.GetHistoryFilter("ai").String(), "ai"},
		{historyfilter.GetHistoryFilter("all").String(), "all"},
		{historyfilter.GetHistoryFilter("").String(), "all"},
		{historyfilter.GetHistoryFilter("other").String(), "unknown"},
		{aiaccessstate.GetAIAccessState("running").String(), "running"},
		{aiaccessstate.GetAIAccessState("error").String(), "error"},
		{aiaccessstate.GetAIAccessState("off").String(), "off"},
		{aiaccessstate.GetAIAccessState("").String(), "off"},
		{aiaccessstate.GetAIAccessState("other").String(), "unknown"},
		{approvaloutcome.GetApprovalOutcome("approved").String(), "approved"},
		{approvaloutcome.GetApprovalOutcome("denied").String(), "denied"},
		{approvaloutcome.GetApprovalOutcome("timed_out").String(), "timed_out"},
		{approvaloutcome.GetApprovalOutcome("pending").String(), "pending"},
		{approvaloutcome.GetApprovalOutcome("").String(), "pending"},
		{approvaloutcome.GetApprovalOutcome("other").String(), "unknown"},
		{aiactivityoutcome.GetAIActivityOutcome(" OK ").String(), "ok"},
		{aiactivityoutcome.GetAIActivityOutcome("error").String(), "error"},
		{aiactivityoutcome.GetAIActivityOutcome("denied").String(), "denied"},
		{aiactivityoutcome.GetAIActivityOutcome("timed_out").String(), "timed_out"},
		{aiactivityoutcome.GetAIActivityOutcome("").String(), "unknown"},
		{aichangekind.GetAIChangeKind("tree").String(), "tree"},
		{aichangekind.GetAIChangeKind("request").String(), "request"},
		{aichangekind.GetAIChangeKind("variables").String(), "variables"},
		{aichangekind.GetAIChangeKind("history").String(), "history"},
		{aichangekind.GetAIChangeKind("cookies").String(), "cookies"},
		{aichangekind.GetAIChangeKind("workspaces").String(), "workspaces"},
		{aichangekind.GetAIChangeKind("git").String(), "unknown"},
		{aitooleffect.GetAIToolEffect(" READ_ONLY ").String(), "read_only"},
		{aitooleffect.GetAIToolEffect("write").String(), "write"},
		{aitooleffect.GetAIToolEffect("destructive").String(), "destructive"},
		{aitooleffect.GetAIToolEffect("other").String(), "unknown"},
	}
	for _, test := range cases {
		if test.got != test.want {
			t.Errorf("got %q, want %q", test.got, test.want)
		}
	}
	if !historysource.AI.Equals("AI") || historykind.Send.Equals("action") || !aiaccessstate.Off.Equals("off") || !aiactivityoutcome.TimedOut.Equals("TIMED_OUT") || !aitooleffect.Destructive.Equals("Destructive") {
		t.Error("Equals must compare case-insensitively")
	}
}
