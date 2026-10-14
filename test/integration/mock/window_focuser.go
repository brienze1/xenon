package mock

import "sync"

type WindowFocuser struct {
	focused int
	mu      sync.RWMutex
}

func NewWindowFocuser() *WindowFocuser {
	return &WindowFocuser{}
}

func (w *WindowFocuser) Focus() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.focused++
}

func (w *WindowFocuser) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.focused
}

func (w *WindowFocuser) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.focused = 0
}
