package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/approvaloutcome"
)

var (
	errUnknownApproval  = errors.New("unknown approval")
	errResolvedApproval = errors.New("already resolved approval")
)

type fakeApprovalQueue struct {
	approvals map[string]entity.AIApproval
	sessions  map[string]string
	resolved  map[string]chan entity.AIApproval
	stale     []entity.AIApproval
	order     []string
	mu        sync.Mutex
}

func newFakeApprovalQueue() *fakeApprovalQueue {
	return &fakeApprovalQueue{approvals: map[string]entity.AIApproval{}, sessions: map[string]string{}, resolved: map[string]chan entity.AIApproval{}}
}

func (f *fakeApprovalQueue) Queue(_ context.Context, approval entity.AIApproval, sessionId string) <-chan entity.AIApproval {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.approvals[approval.ID], f.sessions[approval.ID] = approval, sessionId
	f.resolved[approval.ID] = make(chan entity.AIApproval, 1)
	f.order = append(f.order, approval.ID)
	return f.resolved[approval.ID]
}

func (f *fakeApprovalQueue) FindPending(context.Context) []entity.AIApproval {
	return f.matching(func(string) bool { return true })
}

func (f *fakeApprovalQueue) FindSessionPending(_ context.Context, sessionId string) []entity.AIApproval {
	return f.matching(func(id string) bool { return f.sessions[id] == sessionId })
}

func (f *fakeApprovalQueue) Resolve(_ context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) (*entity.AIApproval, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	approval, found := f.approvals[id]
	if !found {
		return nil, errUnknownApproval
	}
	resolved, accepted := resolve(approval)
	if !accepted {
		return nil, errResolvedApproval
	}
	f.approvals[id] = resolved
	f.resolved[id] <- resolved
	return &resolved, nil
}

func (f *fakeApprovalQueue) Settle(_ context.Context, id string, resolve func(entity.AIApproval) (entity.AIApproval, bool)) (*entity.AIApproval, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	approval, found := f.approvals[id]
	if !found {
		return nil, false
	}
	resolved, accepted := resolve(approval)
	if !accepted {
		return nil, false
	}
	f.approvals[id] = resolved
	f.resolved[id] <- resolved
	return &resolved, true
}

func (f *fakeApprovalQueue) matching(matches func(id string) bool) []entity.AIApproval {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.stale != nil {
		return f.stale
	}
	pending := []entity.AIApproval{}
	for _, id := range f.order {
		if f.approvals[id].IsPending() && matches(id) {
			pending = append(pending, f.approvals[id])
		}
	}
	return pending
}

type fakeApprovalTimer struct {
	expire func(ctx context.Context) error
	delay  time.Duration
	mu     sync.Mutex
}

func (f *fakeApprovalTimer) Schedule(_ context.Context, _ string, delay time.Duration, expire func(ctx context.Context) error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.delay, f.expire = delay, expire
}

type fakeFocus struct{ focused atomic.Int32 }

func (f *fakeFocus) Focus(context.Context) { f.focused.Add(1) }

type fakeApprovalEvents struct {
	requested chan entity.AIApproval
	resolved  []entity.AIApproval
	mu        sync.Mutex
}

func (f *fakeApprovalEvents) AccessStatusChanged(context.Context, entity.AIAccessStatus) error {
	return nil
}

func (f *fakeApprovalEvents) HistoryChanged(context.Context, string) error {
	return nil
}

func (f *fakeApprovalEvents) WorkspacesChanged(context.Context, string) error {
	return nil
}

func (f *fakeApprovalEvents) ApprovalRequested(_ context.Context, approval entity.AIApproval) error {
	f.requested <- approval
	return nil
}

func (f *fakeApprovalEvents) ApprovalResolved(_ context.Context, approval entity.AIApproval) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resolved = append(f.resolved, approval)
	return nil
}

func newApproveAISendFixture() (*approveAISend, *fakeApprovalQueue, *fakeApprovalTimer, *fakeFocus, *fakeApprovalEvents) {
	queue, timer, focus := newFakeApprovalQueue(), &fakeApprovalTimer{}, &fakeFocus{}
	events := &fakeApprovalEvents{requested: make(chan entity.AIApproval, 4)}
	service := NewApproveAISendService(queue, queue, queue, queue, timer, focus, events).(*approveAISend)
	return service, queue, timer, focus, events
}

func requestInBackground(ctx context.Context, service *approveAISend, sessionId string) <-chan entity.AIApproval {
	done := make(chan entity.AIApproval, 1)
	go func() {
		approval, _ := service.Request(ctx, entity.AIApproval{Client: "Claude Code", Method: "DELETE", RawURL: "{{base_url}}/users/7"}, sessionId)
		done <- approval
	}()
	return done
}

func TestApproveAISendWaitsForTheAnswer(t *testing.T) {
	service, queue, timer, focus, events := newApproveAISendFixture()
	done := requestInBackground(context.Background(), service, "s1")
	requested := <-events.requested
	if focus.focused.Load() != 1 || timer.delay != entity.AIApprovalTimeout || requested.Client != "Claude Code" || len(NewFindAIApprovalsService(queue).FindPending(context.Background())) != 1 {
		t.Fatalf("focused=%d delay=%s requested=%+v", focus.focused.Load(), timer.delay, requested)
	}
	if err := service.Answer(context.Background(), requested.ID, true); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if approval := <-done; !approval.IsApproved() {
		t.Fatalf("approval = %+v", approval)
	}
	if err := service.Answer(context.Background(), requested.ID, false); !errors.Is(err, errResolvedApproval) {
		t.Fatalf("second answer = %v", err)
	}
	if err := service.Answer(context.Background(), "unknown", false); !errors.Is(err, errUnknownApproval) {
		t.Fatalf("unknown answer = %v", err)
	}
	if len(events.resolved) != 1 || events.resolved[0].Outcome != approvaloutcome.Approved {
		t.Fatalf("resolved events = %+v", events.resolved)
	}
}

func TestApproveAISendTimesOutDeniesSessionsAndCancellations(t *testing.T) {
	service, queue, timer, _, events := newApproveAISendFixture()
	done := requestInBackground(context.Background(), service, "s1")
	<-events.requested
	if err := timer.expire(context.Background()); err != nil {
		t.Fatalf("expire: %v", err)
	}
	if approval := <-done; approval.Outcome != approvaloutcome.TimedOut {
		t.Fatalf("timed out approval = %+v", approval)
	}

	done = requestInBackground(context.Background(), service, "s2")
	<-events.requested
	if err := service.Deny(context.Background(), "s2"); err != nil {
		t.Fatalf("deny session: %v", err)
	}
	if approval := <-done; approval.Outcome != approvaloutcome.Denied {
		t.Fatalf("session approval = %+v", approval)
	}

	ctx, cancel := context.WithCancel(context.Background())
	done = requestInBackground(ctx, service, "s3")
	<-events.requested
	cancel()
	if approval := <-done; approval.Outcome != approvaloutcome.Denied || len(NewFindAIApprovalsService(queue).FindPending(context.Background())) != 0 {
		t.Fatalf("cancelled approval = %+v", approval)
	}
	if len(events.resolved) != 3 {
		t.Fatalf("resolved events = %d", len(events.resolved))
	}
}

func TestApproveAISendDeniesEverySessionForAnEmptySession(t *testing.T) {
	service, queue, _, _, events := newApproveAISendFixture()
	first, second := requestInBackground(context.Background(), service, "s1"), requestInBackground(context.Background(), service, "s2")
	<-events.requested
	<-events.requested
	if err := service.Deny(context.Background(), ""); err != nil {
		t.Fatalf("deny every session: %v", err)
	}
	if a, b := <-first, <-second; a.Outcome != approvaloutcome.Denied || b.Outcome != approvaloutcome.Denied || len(NewFindAIApprovalsService(queue).FindPending(context.Background())) != 0 {
		t.Fatalf("approvals = %+v, %+v", a, b)
	}
}

func TestApproveAISendTimeoutAfterTheAnswerIsANoOp(t *testing.T) {
	service, _, timer, _, events := newApproveAISendFixture()
	done := requestInBackground(context.Background(), service, "s1")
	requested := <-events.requested
	if err := service.Answer(context.Background(), requested.ID, true); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if approval := <-done; !approval.IsApproved() {
		t.Fatalf("approval = %+v", approval)
	}
	if err := timer.expire(context.Background()); err != nil {
		t.Fatalf("expire after answer: %v", err)
	}
	if len(events.resolved) != 1 || events.resolved[0].Outcome != approvaloutcome.Approved {
		t.Fatalf("resolved events = %+v", events.resolved)
	}
}

func TestApproveAISendDenySkipsAnApprovalResolvedConcurrently(t *testing.T) {
	service, queue, _, _, events := newApproveAISendFixture()
	first, second, third := requestInBackground(context.Background(), service, "s1"), requestInBackground(context.Background(), service, "s1"), requestInBackground(context.Background(), service, "s1")
	answered := <-events.requested
	<-events.requested
	<-events.requested
	queue.stale = NewFindAIApprovalsService(queue).FindPending(context.Background())
	if err := service.Answer(context.Background(), answered.ID, true); err != nil {
		t.Fatalf("answer: %v", err)
	}
	if err := service.Deny(context.Background(), "s1"); err != nil {
		t.Fatalf("deny: %v", err)
	}
	outcomes := map[enums.ApprovalOutcome]int{}
	for _, done := range []<-chan entity.AIApproval{first, second, third} {
		outcomes[(<-done).Outcome]++
	}
	if outcomes[approvaloutcome.Approved] != 1 || outcomes[approvaloutcome.Denied] != 2 || len(events.resolved) != 3 {
		t.Fatalf("outcomes = %+v, resolved events = %d", outcomes, len(events.resolved))
	}
}
