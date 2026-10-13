package publisher

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/publisher/dto"
)

type recordingEmitter struct {
	events []dto.ConsoleEntries
	mutex  sync.Mutex
}

func (r *recordingEmitter) Emit(name string, data any) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if name == dto.ConsoleEntriesEvent {
		r.events = append(r.events, data.(dto.ConsoleEntries))
	}
}

func (r *recordingEmitter) emitted() []dto.ConsoleEntries {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return append([]dto.ConsoleEntries{}, r.events...)
}

func waitForEvents(emitter *recordingEmitter, count int) []dto.ConsoleEntries {
	deadline := time.Now().Add(2 * time.Second)
	for len(emitter.emitted()) < count && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	return emitter.emitted()
}

func entries(count int) []entity.ConsoleEntry {
	result := make([]entity.ConsoleEntry, count)
	for index := range result {
		result[index] = entity.ConsoleEntry{ID: string(rune('a' + index%26))}
	}
	return result
}

func TestConsoleEntriesAreBatched(t *testing.T) {
	ctx := context.Background()
	emitter := &recordingEmitter{}
	publisher := NewConsoleEntries(emitter)
	publisher.PublishEntries(ctx, entries(1))
	publisher.PublishEntries(ctx, entries(2))
	if got := emitter.emitted(); len(got) != 0 {
		t.Fatalf("emitted before the window: %+v", got)
	}
	events := waitForEvents(emitter, 1)
	if len(events) != 1 || len(events[0].Entries) != 3 || events[0].Cleared {
		t.Fatalf("events = %+v", events)
	}
	publisher.PublishEntries(ctx, nil)
	time.Sleep(3 * consoleBatchWindow)
	if got := emitter.emitted(); len(got) != 1 {
		t.Errorf("nothing pending emits nothing: %+v", got)
	}
}

func TestConsoleEntriesFlushAtTheBatchSize(t *testing.T) {
	ctx := context.Background()
	emitter := &recordingEmitter{}
	publisher := NewConsoleEntries(emitter)
	publisher.PublishEntries(ctx, entries(consoleBatchEntries-1))
	if got := emitter.emitted(); len(got) != 0 {
		t.Fatalf("emitted below the batch size: %d", len(got))
	}
	publisher.PublishEntries(ctx, entries(1))
	if got := emitter.emitted(); len(got) != 1 || len(got[0].Entries) != consoleBatchEntries {
		t.Fatalf("a full batch is emitted at once: %+v", got)
	}
	time.Sleep(3 * consoleBatchWindow)
	if got := emitter.emitted(); len(got) != 1 {
		t.Errorf("the flushed batch's timer emitted again: %d", len(got))
	}
}

func TestConsoleClearedDropsPendingEntries(t *testing.T) {
	ctx := context.Background()
	emitter := &recordingEmitter{}
	publisher := NewConsoleEntries(emitter)
	publisher.PublishEntries(ctx, entries(2))
	publisher.PublishCleared(ctx)
	time.Sleep(3 * consoleBatchWindow)
	events := emitter.emitted()
	if len(events) != 1 || !events[0].Cleared || len(events[0].Entries) != 0 || events[0].Entries == nil {
		t.Fatalf("events = %+v", events)
	}
	publisher.PublishEntries(ctx, entries(1))
	if events = waitForEvents(emitter, 2); len(events) != 2 || len(events[1].Entries) != 1 {
		t.Errorf("entries after a clear = %+v", events)
	}
}
