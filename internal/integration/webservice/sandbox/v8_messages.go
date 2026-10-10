package sandbox

import (
	"regexp"
	"strings"
)

var (
	readOfUndefined = regexp.MustCompile(`Cannot read property '([^']*)' of undefined(?: or null)?`)
	noMember        = regexp.MustCompile(`Object has no member '([^']*)'`)
	notIterable     = regexp.MustCompile(`^object is not iterable$`)
	forOfSubject    = regexp.MustCompile(`for\s*\(\s*(?:const|let|var)?\s*[^;()]*?\s+of\s+([A-Za-z_$][\w$.]*)\s*\)`)
)

// v8Message rewrites goja's wording of common TypeErrors into V8's, which Postman scripts and their tests compare against.
func v8Message(text, script string) string {
	if notIterable.MatchString(text) {
		return iterableOf(script) + " is not iterable"
	}
	if !strings.Contains(text, "Cannot read property '") && !strings.Contains(text, "Object has no member '") {
		return text
	}
	text = readOfUndefined.ReplaceAllString(text, "Cannot read properties of undefined (reading '$1')")
	return noMember.ReplaceAllStringFunc(text, func(match string) string {
		return calleeOf(script, noMember.FindStringSubmatch(match)[1]) + " is not a function"
	})
}

func calleeOf(script, member string) string {
	pattern := regexp.MustCompile(`([A-Za-z_$][\w$]*(?:\s*\??\.\s*[A-Za-z_$][\w$]*)*\s*\??\.\s*` + regexp.QuoteMeta(member) + `)\s*\(`)
	found := ""
	for _, match := range pattern.FindAllStringSubmatch(script, -1) {
		callee := strings.Join(strings.Fields(match[1]), "")
		if found != "" && found != callee {
			return member
		}
		found = callee
	}
	if found == "" {
		return member
	}
	return found
}

func iterableOf(script string) string {
	found := ""
	for _, match := range forOfSubject.FindAllStringSubmatch(script, -1) {
		if found != "" && found != match[1] {
			return "object"
		}
		found = match[1]
	}
	if found == "" {
		return "object"
	}
	return found
}
