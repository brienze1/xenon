package webservice

import "strings"

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
