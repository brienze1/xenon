package curl

import (
	"regexp"
	"strings"
)

type dialect int

const (
	bashDialect dialect = iota
	cmdDialect
	powershellDialect
)

const (
	cmdEscape        = '^'
	powershellEscape = '`'
)

var (
	powershellProgram = regexp.MustCompile(`(?i)^\s*curl\.exe\b`)
	cmdMarkers        = []string{"^\n", `^"`}
	powershellMarkers = []string{"`\n"}
)

func normalize(text string) string {
	text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
	switch detectDialect(text) {
	case cmdDialect:
		return toBash(cmdArguments(text))
	case powershellDialect:
		return toBash(powershellArguments(text))
	default:
		return text
	}
}

func detectDialect(text string) dialect {
	switch {
	case powershellProgram.MatchString(text) || containsAny(text, powershellMarkers):
		return powershellDialect
	case containsAny(text, cmdMarkers):
		return cmdDialect
	default:
		return bashDialect
	}
}

func containsAny(text string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func toBash(arguments []string) string {
	quoted := make([]string, 0, len(arguments))
	for index, argument := range arguments {
		if index == 0 {
			argument = programName(argument)
		}
		quoted = append(quoted, "'"+strings.ReplaceAll(argument, "'", `'\''`)+"'")
	}
	return strings.Join(quoted, " ")
}

func programName(argument string) string {
	if strings.EqualFold(argument, "curl.exe") {
		return curlProgram
	}
	return argument
}

func cmdArguments(text string) []string {
	var unescaped strings.Builder
	for position := 0; position < len(text); position++ {
		if text[position] == cmdEscape && position+1 < len(text) {
			position++
			if text[position] == '\n' {
				continue
			}
		}
		unescaped.WriteByte(text[position])
	}
	return windowsArguments(unescaped.String())
}

func windowsArguments(text string) []string {
	parser := &argumentParser{}
	quoted := false
	for position := 0; position < len(text); position++ {
		current := text[position]
		switch {
		case current == '\\':
			position = parser.backslashes(text, position, &quoted)
		case current == '"' && quoted && position+1 < len(text) && text[position+1] == '"':
			parser.write('"')
			position++
		case current == '"':
			quoted = !quoted
			parser.started = true
		case isSpace(current) && !quoted:
			parser.flush()
		default:
			parser.write(current)
		}
	}
	return parser.done()
}

func powershellArguments(text string) []string {
	parser := &argumentParser{}
	var quote byte
	for position := 0; position < len(text); position++ {
		current := text[position]
		next := byte(0)
		if position+1 < len(text) {
			next = text[position+1]
		}
		switch {
		case quote == 0 && current == powershellEscape && next == '\n':
			position++
		case quote != '\'' && current == powershellEscape && next != 0:
			parser.write(powershellEscaped(next))
			position++
		case quote != 0 && current == quote && next == quote:
			parser.write(quote)
			position++
		case quote != 0 && current == quote:
			quote = 0
		case quote == 0 && (current == '\'' || current == '"'):
			quote = current
			parser.started = true
		case quote == 0 && isSpace(current):
			parser.flush()
		default:
			parser.write(current)
		}
	}
	return parser.done()
}

func powershellEscaped(char byte) byte {
	switch char {
	case 'n':
		return '\n'
	case 't':
		return '\t'
	case 'r':
		return '\r'
	default:
		return char
	}
}

func isSpace(char byte) bool {
	return char == ' ' || char == '\t' || char == '\n'
}

type argumentParser struct {
	current   strings.Builder
	arguments []string
	started   bool
}

func (a *argumentParser) write(char byte) {
	a.current.WriteByte(char)
	a.started = true
}

func (a *argumentParser) flush() {
	if a.started {
		a.arguments = append(a.arguments, a.current.String())
	}
	a.current.Reset()
	a.started = false
}

func (a *argumentParser) done() []string {
	a.flush()
	return a.arguments
}

func (a *argumentParser) backslashes(text string, position int, quoted *bool) int {
	count := 0
	for position < len(text) && text[position] == '\\' {
		count++
		position++
	}
	if position < len(text) && text[position] == '"' {
		for range count / 2 {
			a.write('\\')
		}
		if count%2 == 1 {
			a.write('"')
			return position
		}
		*quoted = !*quoted
		a.started = true
		return position
	}
	for range count {
		a.write('\\')
	}
	return position - 1
}
