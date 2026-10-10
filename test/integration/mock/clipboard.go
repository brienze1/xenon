package mock

import "sync"

type Clipboard struct {
	text string
	mu   sync.RWMutex
}

func NewClipboard() *Clipboard {
	return &Clipboard{}
}

func (c *Clipboard) SetText(text string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.text = text
	return nil
}

func (c *Clipboard) Text() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.text
}

func (c *Clipboard) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.text = ""
}
