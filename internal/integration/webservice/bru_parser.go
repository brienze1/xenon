package webservice

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/integration/webservice/bruno/dto"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	blockIndent     = "  "
	multilineMarker = "'''"
	disabledMarker  = "~"
)

var (
	dictionaryHeader = regexp.MustCompile(`^([A-Za-z0-9:_.\-]+)\s*\{\s*$`)
	listHeader       = regexp.MustCompile(`^([A-Za-z0-9:_.\-]+)\s*\[\s*$`)
	textBlocks       = []string{"body:json", "body:text", "body:xml", "body:sparql", "body:graphql", "body:graphql:vars", "script:pre-request", "script:post-response", "tests", "docs"}
)

type bruParser struct {
	file     dto.BruFile
	lines    []string
	position int
}

func parseBru(path string, data []byte) (*dto.BruFile, error) {
	text := strings.ReplaceAll(strings.ReplaceAll(string(data), "\r\n", "\n"), "\r", "\n")
	parser := &bruParser{lines: strings.Split(text, "\n"), file: dto.NewBruFile(path)}
	for parser.position < len(parser.lines) {
		if err := parser.next(); err != nil {
			return nil, err
		}
	}
	return &parser.file, nil
}

func (b *bruParser) next() error {
	line := strings.TrimRight(b.lines[b.position], " \t")
	header := b.position + 1
	b.position++
	switch {
	case strings.TrimSpace(line) == "":
		return nil
	case dictionaryHeader.MatchString(line):
		name := dictionaryHeader.FindStringSubmatch(line)[1]
		if slices.Contains(textBlocks, name) {
			return b.textBlock(name, header)
		}
		return b.dictionaryBlock(name, header)
	case listHeader.MatchString(line):
		return b.listBlock(listHeader.FindStringSubmatch(line)[1], header)
	case strings.Contains(line, ":"):
		entry, err := b.entry(line, header)
		b.file.Attributes = append(b.file.Attributes, entry)
		return err
	default:
		return b.invalid(header, "unexpected line "+strconv.Quote(strings.TrimSpace(line)))
	}
}

func (b *bruParser) textBlock(name string, header int) error {
	content := make([]string, 0)
	for b.position < len(b.lines) {
		line := strings.TrimRight(b.lines[b.position], " \t")
		b.position++
		if line == "}" {
			b.file.Blocks = append(b.file.Blocks, dto.NewTextBruBlock(name, header, strings.Join(content, "\n")))
			return nil
		}
		content = append(content, strings.TrimPrefix(b.lines[b.position-1], blockIndent))
	}
	return b.invalid(header, "the "+name+" block is not closed")
}

func (b *bruParser) dictionaryBlock(name string, header int) error {
	block := dto.NewBruBlock(name, header)
	for b.position < len(b.lines) {
		line := strings.TrimRight(b.lines[b.position], " \t")
		number := b.position + 1
		b.position++
		switch {
		case line == "}":
			b.file.Blocks = append(b.file.Blocks, block)
			return nil
		case strings.TrimSpace(line) == "":
			continue
		}
		entry, err := b.entry(line, number)
		if err != nil {
			return err
		}
		block.Entries = append(block.Entries, entry)
	}
	return b.invalid(header, "the "+name+" block is not closed")
}

func (b *bruParser) listBlock(name string, header int) error {
	block := dto.NewBruBlock(name, header)
	for b.position < len(b.lines) {
		line := strings.TrimSpace(b.lines[b.position])
		number := b.position + 1
		b.position++
		if strings.HasPrefix(line, "]") {
			b.file.Blocks = append(b.file.Blocks, block)
			return nil
		}
		for _, item := range listItems(line) {
			key, enabled := withoutDisabledMarker(item)
			block.Entries = append(block.Entries, dto.NewBruEntry(key, "", number, enabled))
		}
	}
	return b.invalid(header, "the "+name+" list is not closed")
}

func (b *bruParser) entry(line string, number int) (entry dto.BruEntry, err error) {
	key, value, found := splitEntry(strings.TrimSpace(line))
	if !found {
		return entry, b.invalid(number, "expected key: value, got "+strconv.Quote(strings.TrimSpace(line)))
	}
	name, enabled := withoutDisabledMarker(key)
	entry = dto.NewBruEntry(name, value, number, enabled)
	switch {
	case value == multilineMarker:
		return b.multiline(entry)
	case value == "[":
		return b.inlineList(entry)
	case strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]"):
		entry.Values = listItems(value[1 : len(value)-1])
	}
	return entry, nil
}

func (b *bruParser) multiline(entry dto.BruEntry) (dto.BruEntry, error) {
	content := make([]string, 0)
	for b.position < len(b.lines) {
		line := b.lines[b.position]
		b.position++
		if strings.TrimSpace(line) == multilineMarker {
			entry.Value = dedent(content)
			return entry, nil
		}
		content = append(content, strings.TrimRight(line, " \t"))
	}
	return entry, b.invalid(entry.Line, "the multi-line value of "+entry.Key+" is not closed")
}

func (b *bruParser) inlineList(entry dto.BruEntry) (dto.BruEntry, error) {
	entry.Values = []string{}
	for b.position < len(b.lines) {
		line := strings.TrimSpace(b.lines[b.position])
		b.position++
		if line == "]" {
			entry.Value = strings.Join(entry.Values, ", ")
			return entry, nil
		}
		entry.Values = append(entry.Values, listItems(line)...)
	}
	return entry, b.invalid(entry.Line, "the list "+entry.Key+" is not closed")
}

func (b *bruParser) invalid(line int, reason string) error {
	return errs.BadRequestError(b.file.Path+":"+strconv.Itoa(line)+" is not valid Bruno: "+reason, "IMP-36400")
}

func splitEntry(line string) (string, string, bool) {
	if strings.HasPrefix(strings.TrimPrefix(line, disabledMarker), `"`) {
		return splitQuotedEntry(line)
	}
	key, value, found := strings.Cut(line, ":")
	return strings.TrimSpace(key), strings.TrimSpace(value), found && strings.TrimSpace(key) != ""
}

func splitQuotedEntry(line string) (string, string, bool) {
	prefix := ""
	if strings.HasPrefix(line, disabledMarker) {
		prefix, line = disabledMarker, line[1:]
	}
	quoted, err := strconv.QuotedPrefix(line)
	if err != nil {
		return "", "", false
	}
	key, _ := strconv.Unquote(quoted)
	rest := strings.TrimSpace(line[len(quoted):])
	if !strings.HasPrefix(rest, ":") {
		return "", "", false
	}
	return prefix + key, strings.TrimSpace(rest[1:]), true
}

func withoutDisabledMarker(key string) (string, bool) {
	if strings.HasPrefix(key, disabledMarker) {
		return strings.TrimSpace(key[1:]), false
	}
	return key, true
}

func listItems(text string) []string {
	items := make([]string, 0)
	for _, item := range strings.Split(text, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func dedent(lines []string) string {
	indent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if width := len(line) - len(strings.TrimLeft(line, " \t")); indent < 0 || width < indent {
			indent = width
		}
	}
	dedented := make([]string, 0, len(lines))
	for _, line := range lines {
		dedented = append(dedented, line[min(max(indent, 0), len(line)):])
	}
	return strings.Join(dedented, "\n")
}
