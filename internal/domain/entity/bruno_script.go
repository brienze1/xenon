package entity

import "strings"

type BrunoScript struct {
	Text string
	Line int
}

func (b BrunoScript) IsEmpty() bool {
	return strings.TrimSpace(b.Text) == ""
}
