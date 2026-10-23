package entity

import (
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/scriptphase"
)

func TestSendExecutionNextRequestIsTheLastCallAcrossTheChains(t *testing.T) {
	execution := NewSendExecution("send")
	if execution.NextRequest() != nil {
		t.Fatal("no call means no next request")
	}
	pre := NewNextRequest("B")
	post := NewStopNextRequest()
	execution.AddRun(NewScriptRun(ScriptRef{Phase: scriptphase.PreRequest}, time.Time{}).WithEffects(ScriptEffects{NextRequest: &pre}))
	if next := execution.NextRequest(); next == nil || next.Target != "B" || next.Stop {
		t.Errorf("a pre-request call counts: %+v", next)
	}
	execution.AddRun(NewScriptRun(ScriptRef{Phase: scriptphase.PostResponse}, time.Time{}).WithEffects(ScriptEffects{NextRequest: &post}))
	execution.AddRun(NewScriptRun(ScriptRef{Phase: scriptphase.PostResponse}, time.Time{}).WithEffects(ScriptEffects{}))
	if next := execution.NextRequest(); next == nil || !next.Stop {
		t.Errorf("the last call wins and a later script without a call keeps it: %+v", next)
	}
}
