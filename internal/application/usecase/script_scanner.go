package usecase

import "strings"

const pmRoot = "pm"

func pmChains(text string) []pmChain {
	scanner := &scriptScanner{text: text, line: 1}
	return scanner.chains()
}

type scriptScanner struct {
	text     string
	found    []pmChain
	position int
	line     int
	previous byte
}

func (s *scriptScanner) chains() []pmChain {
	for s.position < len(s.text) {
		current := s.text[s.position]
		switch {
		case current == '\n':
			s.line++
			s.advance()
		case strings.HasPrefix(s.text[s.position:], "//"):
			s.skipUntil("\n", false)
		case strings.HasPrefix(s.text[s.position:], "/*"):
			s.skipUntil("*/", true)
		case current == '"' || current == '\'' || current == '`':
			s.skipString(current)
		case isIdentifierStart(current):
			s.identifier()
		default:
			s.advance()
		}
	}
	return s.found
}

func (s *scriptScanner) advance() {
	if s.text[s.position] > ' ' {
		s.previous = s.text[s.position]
	}
	s.position++
}

func (s *scriptScanner) skipUntil(end string, consume bool) {
	closing := strings.Index(s.text[s.position+len(end):], end)
	stop := len(s.text)
	if closing >= 0 {
		stop = s.position + len(end) + closing
		if consume {
			stop += len(end)
		}
	}
	s.line += strings.Count(s.text[s.position:stop], "\n")
	s.position = stop
}

func (s *scriptScanner) skipString(quote byte) {
	s.position++
	for s.position < len(s.text) && s.text[s.position] != quote {
		if s.text[s.position] == '\\' {
			s.position++
		} else if s.text[s.position] == '\n' {
			s.line++
		}
		s.position++
	}
	s.position++
	s.previous = quote
}

func (s *scriptScanner) identifier() {
	start := s.position
	word := s.word()
	if word != pmRoot || s.previous == '.' {
		s.previous = s.text[s.position-1]
		return
	}
	segments := []string{word}
	for s.position+1 < len(s.text) && s.text[s.position] == '.' && isIdentifierStart(s.text[s.position+1]) {
		s.position++
		segments = append(segments, s.word())
	}
	s.previous = s.text[s.position-1]
	if s.position > start+len(pmRoot) {
		s.found = append(s.found, pmChain{chain: strings.Join(segments, "."), line: s.line})
	}
}

func (s *scriptScanner) word() string {
	start := s.position
	for s.position < len(s.text) && isIdentifierPart(s.text[s.position]) {
		s.position++
	}
	return s.text[start:s.position]
}
