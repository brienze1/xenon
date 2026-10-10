package usecase

import "strings"

type leftoverScanner struct {
	text     string
	flagged  []bool
	position int
	line     int
}

func (l *leftoverScanner) scan() {
	for l.position < len(l.text) {
		rest := l.text[l.position:]
		switch current := rest[0]; {
		case current == '\n':
			l.line++
			l.position++
		case strings.HasPrefix(rest, "//"):
			l.skip(strings.IndexByte(rest, '\n'))
		case strings.HasPrefix(rest, "/*"):
			l.skip(blockCommentLength(rest))
		case current == '"' || current == '\'':
			l.skip(skipQuoted(rest, 0, current) + 1)
		case isIdentifierStart(current):
			l.identifier()
		default:
			l.position++
		}
	}
}

func (l *leftoverScanner) skip(length int) {
	if length <= 0 {
		length = len(l.text) - l.position
	}
	end := min(l.position+length, len(l.text))
	l.line += strings.Count(l.text[l.position:end], "\n")
	l.position = end
}

func (l *leftoverScanner) identifier() {
	member := l.position > 0 && l.text[l.position-1] == '.'
	start := l.position
	for l.position < len(l.text) && isIdentifierPart(l.text[l.position]) {
		l.position++
	}
	name := l.text[start:l.position]
	next := byte(0)
	if l.position < len(l.text) {
		next = l.text[l.position]
	}
	if member {
		return
	}
	if (isBrunoObject(name) && (next == '.' || next == '(')) || (name == "require" && next == '(') {
		l.flagged[l.line] = true
	}
}
