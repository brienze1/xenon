package usecase

import (
	"sort"
	"strconv"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/pmapi"
	"github.com/brienze1/xenon/internal/domain/enums/warningcode"
)

const (
	pmRoot             = "pm"
	reportedPmSegments = 3
)

type pmChain struct {
	chain string
	line  int
}

func ScanPmUsage(scripts []entity.ScriptRef) []entity.PmUsage {
	occurrences := map[string][]entity.PmOccurrence{}
	for _, script := range scripts {
		for _, found := range pmChains(script.Text) {
			if isInMvpSubset(found.chain) {
				continue
			}
			api := reportedApi(found.chain)
			occurrences[api] = append(occurrences[api], entity.PmOccurrence{ItemID: script.ItemID, Path: script.Path, Phase: script.Phase, Line: found.line})
		}
	}
	apis := make([]string, 0, len(occurrences))
	for api := range occurrences {
		apis = append(apis, api)
	}
	sort.Strings(apis)
	usages := make([]entity.PmUsage, 0, len(apis))
	for _, api := range apis {
		usages = append(usages, entity.PmUsage{API: api, Occurrences: occurrences[api]})
	}
	return usages
}

func PmUsageWarnings(usages []entity.PmUsage) []entity.ImportWarning {
	warnings := make([]entity.ImportWarning, 0)
	seen := map[string]bool{}
	for _, usage := range usages {
		for _, occurrence := range usage.Occurrences {
			key := usage.API + "\x00" + occurrence.ItemID + "\x00" + occurrence.Phase.String()
			if seen[key] {
				continue
			}
			seen[key] = true
			location := occurrence.Phase.String() + ":" + strconv.Itoa(occurrence.Line)
			warnings = append(warnings, pmUsageWarning(usage.API, occurrence).AtLine(location))
		}
	}
	return warnings
}

func pmUsageWarning(api string, occurrence entity.PmOccurrence) entity.ImportWarning {
	if api == pmapi.Require.String() {
		return entity.NewImportWarning(warningcode.PmRequire, "The script uses pm.require, which only works inside Postman", occurrence.ItemID, occurrence.Path)
	}
	return entity.NewImportWarning(warningcode.PmApiOutsideMvp, "The script uses "+api+", which is outside Xenon's script subset", occurrence.ItemID, occurrence.Path)
}

func isInMvpSubset(chain string) bool {
	for _, allowed := range pmapi.MvpSubset {
		api := allowed.String()
		if chain == api || strings.HasPrefix(chain, api+".") || strings.HasPrefix(api, chain+".") {
			return true
		}
	}
	return false
}

func reportedApi(chain string) string {
	segments := strings.Split(chain, ".")
	if len(segments) > reportedPmSegments {
		segments = segments[:reportedPmSegments]
	}
	return strings.Join(segments, ".")
}

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

func isIdentifierStart(char byte) bool {
	return char == '_' || char == '$' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')
}

func isIdentifierPart(char byte) bool {
	return isIdentifierStart(char) || (char >= '0' && char <= '9')
}
