package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums/brunoruleform"
)

type brunoTranslator struct {
	text     string
	out      strings.Builder
	position int
}

func (b *brunoTranslator) code(stopAtBrace bool) {
	depth := 0
	for b.position < len(b.text) {
		current := b.text[b.position]
		rest := b.text[b.position:]
		switch {
		case strings.HasPrefix(rest, "//"):
			b.copyUntil("\n", false)
		case strings.HasPrefix(rest, "/*"):
			b.copyUntil("*/", true)
		case current == '"' || current == '\'':
			b.copyString(current)
		case current == '`':
			b.template()
		case current == '{':
			depth++
			b.copyByte()
		case current == '}' && depth == 0 && stopAtBrace:
			return
		case current == '}':
			depth--
			b.copyByte()
		case isIdentifierStart(current) && !b.isMemberAccess():
			b.identifier()
		case isIdentifierStart(current):
			b.out.WriteString(b.word())
		default:
			b.copyByte()
		}
	}
}

func (b *brunoTranslator) copyByte() {
	b.out.WriteByte(b.text[b.position])
	b.position++
}

func (b *brunoTranslator) copyUntil(end string, inclusive bool) {
	stop := strings.Index(b.text[b.position+len(end):], end)
	if stop < 0 {
		b.out.WriteString(b.text[b.position:])
		b.position = len(b.text)
		return
	}
	stop += b.position + len(end)
	if inclusive {
		stop += len(end)
	}
	b.out.WriteString(b.text[b.position:stop])
	b.position = stop
}

func (b *brunoTranslator) copyString(quote byte) {
	start := b.position
	b.position++
	for b.position < len(b.text) && b.text[b.position] != quote && b.text[b.position] != '\n' {
		if b.text[b.position] == '\\' {
			b.position++
		}
		b.position++
	}
	b.position = min(b.position+1, len(b.text))
	b.out.WriteString(b.text[start:b.position])
}

func (b *brunoTranslator) template() {
	b.copyByte()
	for b.position < len(b.text) {
		rest := b.text[b.position:]
		switch {
		case rest[0] == '\\' && len(rest) > 1:
			b.out.WriteString(rest[:2])
			b.position += 2
		case strings.HasPrefix(rest, "${"):
			b.out.WriteString("${")
			b.position += 2
			b.code(true)
			if b.position < len(b.text) {
				b.copyByte()
			}
		case rest[0] == '`':
			b.copyByte()
			return
		default:
			b.copyByte()
		}
	}
}

func (b *brunoTranslator) isMemberAccess() bool {
	return b.position > 0 && b.text[b.position-1] == '.'
}

func (b *brunoTranslator) word() string {
	start := b.position
	for b.position < len(b.text) && isIdentifierPart(b.text[b.position]) {
		b.position++
	}
	return b.text[start:b.position]
}

func (b *brunoTranslator) identifier() {
	start := b.position
	name := b.word()
	if isBrunoObject(name) && b.position+1 < len(b.text) && b.text[b.position] == '.' && isIdentifierStart(b.text[b.position+1]) {
		b.position++
		name += "." + b.word()
	}
	rule, found := brunoRules[name]
	if !found || !b.apply(rule) {
		b.position = start
		b.out.WriteString(b.word())
	}
}

func (b *brunoTranslator) apply(rule brunoRule) bool {
	rest := b.text[b.position:]
	switch rule.form {
	case brunoruleform.RenamedCall:
		if !strings.HasPrefix(rest, "(") {
			return false
		}
		b.out.WriteString(rule.replacement)
	case brunoruleform.EmptyCall:
		if !strings.HasPrefix(rest, "()") {
			return false
		}
		b.out.WriteString(rule.replacement)
		b.position += len("()")
	case brunoruleform.Property:
		if strings.HasPrefix(rest, "(") {
			return false
		}
		b.out.WriteString(rule.replacement)
	case brunoruleform.WrappedCall:
		return b.wrap(rule)
	}
	return true
}

func (b *brunoTranslator) wrap(rule brunoRule) bool {
	arguments, end, found := callArguments(b.text, b.position)
	if !found || len(arguments) != rule.arity {
		return false
	}
	parts := strings.Split(rule.replacement, "%s")
	b.out.WriteString(parts[0])
	for index, argument := range arguments {
		b.out.WriteString(translateBrunoCode(strings.Trim(argument, " \t")))
		b.out.WriteString(parts[index+1])
	}
	b.position = end
	return true
}
