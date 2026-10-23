package entity

import (
	"strconv"
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/approvaloutcome"
)

const AIApprovalTimeout = 120 * time.Second

type AIApproval struct {
	CreatedAt   time.Time
	Environment AIApprovalEnvironment
	ID          string
	Client      string
	Method      string
	RawURL      string
	Summary     string
	Outcome     enums.ApprovalOutcome
}

func NewAIApproval(id, client, method, rawURL string, environment AIApprovalEnvironment, createdAt time.Time) AIApproval {
	return AIApproval{
		ID:          id,
		Client:      AIClientName(client),
		Method:      method,
		RawURL:      rawURL,
		Environment: environment,
		CreatedAt:   createdAt,
		Outcome:     approvaloutcome.Pending,
	}
}

func (a AIApproval) Queued(id string, createdAt time.Time) AIApproval {
	queued := NewAIApproval(id, a.Client, a.Method, a.RawURL, a.Environment, createdAt)
	queued.Summary = a.Summary
	return queued
}

func NewRunApprovalSummary(targetName string, requests int) string {
	noun := " requests"
	if requests == 1 {
		noun = " request"
	}
	return "Run " + targetName + " · " + strconv.Itoa(requests) + noun
}

const EverySessionID = ""

func IsEverySession(sessionId string) bool {
	return sessionId == EverySessionID
}

func (a AIApproval) IsPending() bool {
	return a.Outcome == approvaloutcome.Pending
}

func (a AIApproval) IsApproved() bool {
	return a.Outcome == approvaloutcome.Approved
}

func (a AIApproval) IsTimedOut() bool {
	return a.Outcome == approvaloutcome.TimedOut
}

func (a AIApproval) Denied() (AIApproval, bool) {
	return a.Resolved(approvaloutcome.Denied)
}

func (a AIApproval) TimedOut() (AIApproval, bool) {
	return a.Resolved(approvaloutcome.TimedOut)
}

// Resolved reports false when the approval was already answered; every terminal outcome is final.
func (a AIApproval) Resolved(outcome enums.ApprovalOutcome) (AIApproval, bool) {
	if !a.IsPending() || outcome == approvaloutcome.Pending {
		return a, false
	}
	a.Outcome = outcome
	return a, true
}

func (a AIApproval) Answered(approved bool) (AIApproval, bool) {
	if approved {
		return a.Resolved(approvaloutcome.Approved)
	}
	return a.Resolved(approvaloutcome.Denied)
}

func (a AIApproval) LogValue() any {
	return map[string]any{"id": a.ID, "client": a.Client, "method": a.Method, "environmentId": a.Environment.ID, "outcome": a.Outcome}
}
