package entity

import "strings"

type BrunoAssertion struct {
	Expression string
	Operator   string
	Value      string
	Line       int
	Enabled    bool
}

func (b BrunoAssertion) Source() string {
	return strings.TrimSpace(b.Expression + ": " + strings.TrimSpace(b.Operator+" "+b.Value))
}

func (b BrunoAssertion) Title() string {
	return strings.TrimSpace(b.Expression + " " + strings.TrimSpace(b.Operator+" "+b.Value))
}
