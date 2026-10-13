package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ScriptSnippet struct {
	Phase enums.SnippetPhase
	ID    string
	Title string
	Text  string
}
