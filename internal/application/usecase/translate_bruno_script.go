package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const untranslatedPrefix = "// XENON: untranslated "

var brunoObjects = []string{"bru", "req", "res"}

func TranslateBrunoScript(script entity.BrunoScript) entity.ScriptTranslation {
	translation := entity.NewScriptTranslation()
	if strings.TrimSpace(script.Text) == "" {
		return translation
	}
	originals := strings.Split(script.Text, "\n")
	code := translateBrunoCode(script.Text)
	translated := strings.Split(code, "\n")
	flagged := untranslatedLines(code)
	for index, line := range translated {
		if !flagged[index] {
			translation.Lines = append(translation.Lines, line)
			continue
		}
		translation.Lines = append(translation.Lines, commentedOut(originals[index]))
		translation.Untranslated = append(translation.Untranslated, entity.UntranslatedLine{Line: script.Line + index, Source: strings.TrimSpace(originals[index])})
	}
	return translation
}

func commentedOut(line string) string {
	indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	return indent + untranslatedPrefix + strings.TrimSpace(line)
}

func translateBrunoExpression(expression string) (string, bool) {
	translated := translateBrunoCode(strings.TrimSpace(expression))
	for _, flagged := range untranslatedLines(translated) {
		if flagged {
			return translated, false
		}
	}
	return translated, true
}

func translateBrunoCode(code string) string {
	translator := &brunoTranslator{text: code}
	translator.code(false)
	return translator.out.String()
}

func callArguments(text string, open int) ([]string, int, bool) {
	if open >= len(text) || text[open] != '(' {
		return nil, open, false
	}
	arguments := make([]string, 0)
	depth := 0
	start := open + 1
	for position := open; position < len(text); position++ {
		switch current := text[position]; {
		case current == '"' || current == '\'' || current == '`':
			position = skipQuoted(text, position, current)
		case current == '(' || current == '[' || current == '{':
			depth++
		case current == ',' && depth == 1:
			arguments = append(arguments, text[start:position])
			start = position + 1
		case current == ')' && depth == 1:
			if strings.TrimSpace(text[start:position]) != "" || len(arguments) > 0 {
				arguments = append(arguments, text[start:position])
			}
			return arguments, position + 1, true
		case current == ')' || current == ']' || current == '}':
			depth--
		}
	}
	return nil, open, false
}

func skipQuoted(text string, position int, quote byte) int {
	for position++; position < len(text) && text[position] != quote; position++ {
		if text[position] == '\\' {
			position++
		}
	}
	return position
}

func isBrunoObject(name string) bool {
	for _, object := range brunoObjects {
		if name == object {
			return true
		}
	}
	return false
}

func untranslatedLines(code string) []bool {
	flagged := make([]bool, strings.Count(code, "\n")+1)
	scanner := &leftoverScanner{text: code, flagged: flagged}
	scanner.scan()
	return flagged
}

func blockCommentLength(text string) int {
	end := strings.Index(text, "*/")
	if end < 0 {
		return len(text)
	}
	return end + len("*/")
}
