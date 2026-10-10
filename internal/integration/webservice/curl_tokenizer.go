package webservice

import (
	"strconv"
	"strings"

	"mvdan.cc/sh/v3/syntax"

	"github.com/brienze1/xenon/pkg/errs"
)

const curlProgram = "curl"

var (
	ansiCSimple = map[byte]string{'n': "\n", 't': "\t", 'r': "\r", 'a': "\a", 'b': "\b", 'f': "\f", 'v': "\v", 'e': "\x1b", '\\': "\\", '\'': "'", '"': "\""}
	ansiCWidths = map[byte]int{'x': 2, 'u': 4, 'U': 8}
)

func tokenizeCurl(text string) ([]string, error) {
	file, err := syntax.NewParser(syntax.Variant(syntax.LangBash)).Parse(strings.NewReader(text), "")
	if err != nil {
		return nil, errs.BadRequestError("The cURL command could not be read", "IMP-06400")
	}
	for _, statement := range file.Stmts {
		call, isCall := statement.Cmd.(*syntax.CallExpr)
		if !isCall || len(call.Args) == 0 {
			continue
		}
		arguments := make([]string, 0, len(call.Args))
		for _, word := range call.Args {
			arguments = append(arguments, wordText(word, text))
		}
		if !strings.EqualFold(programName(arguments[0]), curlProgram) {
			break
		}
		return arguments[1:], nil
	}
	return nil, errs.BadRequestError("The text is not a cURL command", "IMP-06400")
}

func wordText(word *syntax.Word, source string) string {
	var text strings.Builder
	for _, part := range word.Parts {
		appendPart(&text, part, source, false)
	}
	return text.String()
}

func appendPart(text *strings.Builder, part syntax.WordPart, source string, quoted bool) {
	switch typed := part.(type) {
	case *syntax.Lit:
		text.WriteString(unescapeLiteral(typed.Value, quoted))
	case *syntax.SglQuoted:
		text.WriteString(singleQuoted(typed))
	case *syntax.DblQuoted:
		for _, inner := range typed.Parts {
			appendPart(text, inner, source, true)
		}
	default:
		text.WriteString(source[part.Pos().Offset():part.End().Offset()])
	}
}

func singleQuoted(quoted *syntax.SglQuoted) string {
	if quoted.Dollar {
		return ansiC(quoted.Value)
	}
	return quoted.Value
}

func unescapeLiteral(value string, quoted bool) string {
	var text strings.Builder
	for position := 0; position < len(value); position++ {
		if value[position] != '\\' || position+1 == len(value) {
			text.WriteByte(value[position])
			continue
		}
		next := value[position+1]
		switch {
		case next == '\n':
			position++
		case !quoted || strings.IndexByte("$`\"\\", next) >= 0:
			text.WriteByte(next)
			position++
		default:
			text.WriteByte('\\')
		}
	}
	return text.String()
}

func ansiC(value string) string {
	var text strings.Builder
	for position := 0; position < len(value); position++ {
		if value[position] != '\\' || position+1 == len(value) {
			text.WriteByte(value[position])
			continue
		}
		position++
		position = ansiCEscape(&text, value, position)
	}
	return text.String()
}

func ansiCEscape(text *strings.Builder, value string, position int) int {
	escape := value[position]
	if replacement, found := ansiCSimple[escape]; found {
		text.WriteString(replacement)
		return position
	}
	if width, found := ansiCWidths[escape]; found && position+width < len(value) {
		if code, err := strconv.ParseUint(value[position+1:position+1+width], 16, 32); err == nil {
			writeCode(text, escape, code)
			return position + width
		}
	}
	text.WriteByte('\\')
	text.WriteByte(escape)
	return position
}

func writeCode(text *strings.Builder, escape byte, code uint64) {
	if escape == 'x' {
		text.WriteByte(byte(code))
		return
	}
	text.WriteRune(rune(code))
}
