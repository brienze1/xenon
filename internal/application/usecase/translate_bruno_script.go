package usecase

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const untranslatedPrefix = "// XENON: untranslated "

type brunoRuleForm int

const (
	renamedCall brunoRuleForm = iota
	emptyCall
	property
	wrappedCall
)

type brunoRule struct {
	replacement string
	form        brunoRuleForm
	arity       int
}

var brunoRules = map[string]brunoRule{
	"bru.getEnvVar":        {replacement: "pm.environment.get", form: renamedCall},
	"bru.setEnvVar":        {replacement: "pm.environment.set", form: renamedCall},
	"bru.getVar":           {replacement: "pm.variables.get", form: renamedCall},
	"bru.setVar":           {replacement: "pm.variables.set", form: renamedCall},
	"bru.getRequestVar":    {replacement: "pm.variables.get", form: renamedCall},
	"bru.getCollectionVar": {replacement: "pm.collectionVariables.get", form: renamedCall},
	"bru.getGlobalEnvVar":  {replacement: "pm.globals.get", form: renamedCall},
	"bru.setGlobalEnvVar":  {replacement: "pm.globals.set", form: renamedCall},
	"bru.getEnvName":       {replacement: "pm.environment.name", form: emptyCall},
	"bru.interpolate":      {replacement: "pm.variables.replaceIn", form: renamedCall},
	"req.getUrl":           {replacement: "pm.request.url.toString()", form: emptyCall},
	"req.url":              {replacement: "pm.request.url.toString()", form: property},
	"req.getMethod":        {replacement: "pm.request.method", form: emptyCall},
	"req.method":           {replacement: "pm.request.method", form: property},
	"req.getHeader":        {replacement: "pm.request.headers.get", form: renamedCall},
	"req.setHeader":        {replacement: "pm.request.headers.upsert({ key: %s, value: %s })", form: wrappedCall, arity: 2},
	"req.getHeaders":       {replacement: "pm.request.headers.toObject()", form: emptyCall},
	"req.headers":          {replacement: "pm.request.headers.toObject()", form: property},
	"req.getBody":          {replacement: "JSON.parse(pm.request.body.raw)", form: emptyCall},
	"req.body":             {replacement: "JSON.parse(pm.request.body.raw)", form: property},
	"req.setBody":          {replacement: "pm.request.body.update(JSON.stringify(%s))", form: wrappedCall, arity: 1},
	"res.getStatus":        {replacement: "pm.response.code", form: emptyCall},
	"res.status":           {replacement: "pm.response.code", form: property},
	"res.getBody":          {replacement: "pm.response.json()", form: emptyCall},
	"res.body":             {replacement: "pm.response.json()", form: property},
	"res.getHeader":        {replacement: "pm.response.headers.get", form: renamedCall},
	"res.getHeaders":       {replacement: "pm.response.headers.toObject()", form: emptyCall},
	"res.headers":          {replacement: "pm.response.headers.toObject()", form: property},
	"res.getResponseTime":  {replacement: "pm.response.responseTime", form: emptyCall},
	"res.responseTime":     {replacement: "pm.response.responseTime", form: property},
	"test":                 {replacement: "pm.test", form: renamedCall},
	"expect":               {replacement: "pm.expect", form: renamedCall},
}

var brunoObjects = []string{"bru", "req", "res"}

func TranslateBrunoScript(script entity.BrunoScript) entity.ScriptTranslation {
	translation := entity.ScriptTranslation{Lines: []string{}, Untranslated: []entity.UntranslatedLine{}}
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
	case renamedCall:
		if !strings.HasPrefix(rest, "(") {
			return false
		}
		b.out.WriteString(rule.replacement)
	case emptyCall:
		if !strings.HasPrefix(rest, "()") {
			return false
		}
		b.out.WriteString(rule.replacement)
		b.position += len("()")
	case property:
		if strings.HasPrefix(rest, "(") {
			return false
		}
		b.out.WriteString(rule.replacement)
	case wrappedCall:
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
		b.out.WriteString(translateBrunoCode(strings.TrimSpace(argument)))
		b.out.WriteString(parts[index+1])
	}
	b.position = end
	return true
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

func blockCommentLength(text string) int {
	end := strings.Index(text, "*/")
	if end < 0 {
		return len(text)
	}
	return end + len("*/")
}
