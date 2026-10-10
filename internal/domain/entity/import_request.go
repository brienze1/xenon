package entity

import "strings"

type ImportRequest struct {
	Text  string
	Paths []string
}

func (i ImportRequest) HasInput() bool {
	return len(i.Paths) > 0 || strings.TrimSpace(i.Text) != ""
}

func (i ImportRequest) IsText() bool {
	return len(i.Paths) == 0
}
