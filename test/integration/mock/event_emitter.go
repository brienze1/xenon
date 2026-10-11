package mock

import (
	"sync"
	"time"
)

type EventEmitter struct {
	hooks  []emitHook
	events []emittedEvent
	mu     sync.RWMutex
}

type emitHook struct {
	action func()
	name   string
	count  int
}

type emittedEvent struct {
	at   time.Time
	data any
	name string
}

func NewEventEmitter() *EventEmitter {
	return &EventEmitter{}
}

func (e *EventEmitter) Emit(name string, data any) {
	e.mu.Lock()
	e.events = append(e.events, emittedEvent{name: name, data: data, at: time.Now()})
	actions := e.triggered(name)
	e.mu.Unlock()
	for _, action := range actions {
		action()
	}
}

func (e *EventEmitter) OnEmit(name string, count int, action func()) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.hooks = append(e.hooks, emitHook{name: name, count: count, action: action})
}

func (e *EventEmitter) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.events = nil
	e.hooks = nil
}

func (e *EventEmitter) triggered(name string) []func() {
	count := 0
	for _, event := range e.events {
		if event.name == name {
			count++
		}
	}
	actions := make([]func(), 0)
	for _, hook := range e.hooks {
		if hook.name == name && hook.count == count {
			actions = append(actions, hook.action)
		}
	}
	return actions
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

func (e *EventEmitter) Gaps(name string) []time.Duration {
	e.mu.RLock()
	defer e.mu.RUnlock()
	gaps := make([]time.Duration, 0)
	var previous time.Time
	for _, event := range e.events {
		if event.name != name {
			continue
		}
		if !previous.IsZero() {
			gaps = append(gaps, event.at.Sub(previous))
		}
		previous = event.at
	}
	return gaps
}
