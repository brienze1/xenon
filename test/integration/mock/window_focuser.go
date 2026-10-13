package mock

import "sync"

type WindowFocuser struct {
	focused int
	mu      sync.Mutex
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
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.focused
}

func (w *WindowFocuser) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.focused = 0
}
