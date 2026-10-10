package mock

import (
	"bytes"
	"sync"
)

type LogOutput struct {
	buffer bytes.Buffer
	mu     sync.RWMutex
}

func NewLogOutput() *LogOutput {
	return &LogOutput{}
}

func (l *LogOutput) Write(data []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buffer.Write(data)
}

func (l *LogOutput) Contains(text string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return bytes.Contains(l.buffer.Bytes(), []byte(text))
}

func (l *LogOutput) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buffer.Reset()
}
