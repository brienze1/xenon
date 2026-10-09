package mock

import "sync"

type EventEmitter struct {
	events []emittedEvent
	mu     sync.RWMutex
}

type emittedEvent struct {
	data any
	name string
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{}
}

func (e *EventEmitter) Emit(name string, data any) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = append(e.events, emittedEvent{name: name, data: data})
}

func (e *EventEmitter) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = nil
}

func (e *EventEmitter) Count(name string) int {
	return len(e.named(name))
}

func (e *EventEmitter) Event(name string, index int) (any, bool) {
	events := e.named(name)
	if index < 1 || index > len(events) {
		return nil, false
	}
	return events[index-1], true
}

func (e *EventEmitter) named(name string) []any {
	e.mu.RLock()
	defer e.mu.RUnlock()
	matching := make([]any, 0)
	for _, event := range e.events {
		if event.name == name {
			matching = append(matching, event.data)
		}
	}
	return matching
}
