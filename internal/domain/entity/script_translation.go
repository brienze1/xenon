package entity

import "strings"

type UntranslatedLine struct {
	Source string
	Line   int
}

type ScriptTranslation struct {
	Lines        []string
	Untranslated []UntranslatedLine
}

func (s ScriptTranslation) Append(other ScriptTranslation) ScriptTranslation {
	return ScriptTranslation{
		Lines:        append(append([]string{}, s.Lines...), other.Lines...),
		Untranslated: append(append([]UntranslatedLine{}, s.Untranslated...), other.Untranslated...),
	}
}

func (s ScriptTranslation) Text() string {
	return strings.Join(s.Lines, "\n")
}
