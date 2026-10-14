package entity

import (
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums/approvaloutcome"
)

func TestAIApprovalResolvesOnlyOnce(t *testing.T) {
	created := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	pending := NewAIApproval("01A", "", "GET", "{{api}}/ping", AIApprovalEnvironment{ID: "e", Name: "Prod", Color: "red"}, created)
	if !pending.IsPending() || pending.Client != DefaultAIClientName || pending.RawURL != "{{api}}/ping" {
		t.Fatalf("pending = %+v", pending)
	}
	cases := []struct {
		name     string
		resolve  func(AIApproval) (AIApproval, bool)
		outcome  string
		accepted bool
	}{
		{name: "approve", resolve: func(a AIApproval) (AIApproval, bool) { return a.Answered(true) }, outcome: "approved", accepted: true},
		{name: "deny", resolve: func(a AIApproval) (AIApproval, bool) { return a.Answered(false) }, outcome: "denied", accepted: true},
		{name: "time out", resolve: func(a AIApproval) (AIApproval, bool) { return a.Resolved(approvaloutcome.TimedOut) }, outcome: "timed_out", accepted: true},
		{name: "back to pending", resolve: func(a AIApproval) (AIApproval, bool) { return a.Resolved(approvaloutcome.Pending) }, outcome: "pending", accepted: false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			resolved, accepted := test.resolve(pending)
			if accepted != test.accepted || resolved.Outcome.String() != test.outcome {
				t.Fatalf("resolved = %+v, %v", resolved, accepted)
			}
			if !accepted {
				return
			}
			if again, second := resolved.Answered(true); second || again.Outcome != resolved.Outcome {
				t.Errorf("a second answer must be refused: %+v", again)
			}
		})
	}
	approved, _ := pending.Answered(true)
	if !approved.IsApproved() || pending.IsApproved() {
		t.Error("IsApproved")
	}
	if logged := pending.LogValue().(map[string]any); logged["environmentId"] != "e" {
		t.Errorf("LogValue() = %v", logged)
	}
}

func TestAIApprovalQueuedStartsAFreshPendingApproval(t *testing.T) {
	created := time.Date(2026, 10, 13, 10, 0, 0, 0, time.UTC)
	environment := AIApprovalEnvironment{ID: "e", Name: "Prod", Color: "red"}
	request := AIApproval{ID: "stale", Client: "", Method: "DELETE", RawURL: "{{api}}/users/7", Environment: environment, Outcome: approvaloutcome.Denied}
	queued := request.Queued("01B", created)
	want := NewAIApproval("01B", "", "DELETE", "{{api}}/users/7", environment, created)
	if queued != want {
		t.Fatalf("Queued() = %+v, want %+v", queued, want)
	}
	if !queued.IsPending() || queued.Client != DefaultAIClientName {
		t.Errorf("queued = %+v", queued)
	}
}
