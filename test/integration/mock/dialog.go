package mock

import (
	"strings"
	"sync"
)

type Dialog struct {
	answers        []string
	titles         []string
	suggestedNames []string
	filters        [][]string
	mu             sync.Mutex
}

func NewDialog() *Dialog {
	return &Dialog{}
}

func (d *Dialog) QueueAnswer(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.answers = append(d.answers, path)
}

func (d *Dialog) PickFolder(title string) (string, error) {
	return d.answer(title)
}

func (d *Dialog) PickFile(title string) (string, error) {
	return d.answer(title)
}

func (d *Dialog) PickSources(title string) ([]string, error) {
	answer, err := d.answer(title)
	if answer == "" {
		return []string{}, err
	}
	return strings.Split(answer, "|"), err
}

func (d *Dialog) PickSaveFile(title, suggestedName string) (string, error) {
	d.mu.Lock()
	d.suggestedNames = append(d.suggestedNames, suggestedName)
	d.mu.Unlock()
	return d.answer(title)
}

func (d *Dialog) LastSuggestedName() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.suggestedNames) == 0 {
		return ""
	}
	return d.suggestedNames[len(d.suggestedNames)-1]
}

func (d *Dialog) answer(title string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.titles = append(d.titles, title)
	if len(d.answers) == 0 {
		return "", nil
	}
	answer := d.answers[0]
	d.answers = d.answers[1:]
	return answer, nil
}

func (d *Dialog) LastTitle() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.titles) == 0 {
		return ""
	}
	return d.titles[len(d.titles)-1]
}

func (d *Dialog) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.answers = nil
	d.titles = nil
	d.suggestedNames = nil
	d.filters = nil
}

func (d *Dialog) RecordFilters(filters []string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.filters = append(d.filters, append([]string{}, filters...))
}

func (d *Dialog) LastFilters() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.filters) == 0 {
		return nil
	}
	return d.filters[len(d.filters)-1]
}
