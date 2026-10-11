package snippets

import (
	"strings"
)

const (
	upperHex             = "0123456789ABCDEF"
	componentSafeSymbols = "-_.!~*'()"
)

var paramDecodes = strings.NewReplacer("%5B", "[", "%7B", "{", "%5D", "]", "%7D", "}", "%2B", "+", "%25", "%", "'", "%27")

var escapedSequences = strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`)

func encodeURIComponent(text string) string {
	var builder strings.Builder
	for index := 0; index < len(text); index++ {
		character := text[index]
		if isAlphanumeric(character) || strings.IndexByte(componentSafeSymbols, character) >= 0 {
			builder.WriteByte(character)
			continue
		}
		builder.WriteByte('%')
		builder.WriteByte(upperHex[character>>4])
		builder.WriteByte(upperHex[character&0x0f])
	}
	return builder.String()
}

func isAlphanumeric(character byte) bool {
	return character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z' || character >= '0' && character <= '9'
}

func encodeParam(text string) string {
	return paramDecodes.Replace(encodeURIComponent(text))
}

func encodedQueryUrl(raw string) string {
	rest, hash, hasHash := strings.Cut(raw, "#")
	base, query, _ := strings.Cut(rest, "?")
	pairs := make([]string, 0)
	for _, pair := range strings.Split(query, "&") {
		if pair == "" {
			continue
		}
		key, value, hasValue := strings.Cut(pair, "=")
		if !hasValue {
			pairs = append(pairs, encodeParam(key))
			continue
		}
		pairs = append(pairs, encodeParam(key)+"="+encodeParam(value))
	}
	result := base
	if len(pairs) > 0 {
		result += "?" + strings.Join(pairs, "&")
	}
	if hasHash && hash != "" {
		result += "#" + hash
	}
	return result
}

func escapeString(text string, trim bool) string {
	return trimIf(escapedSequences.Replace(text), trim)
}

func trimIf(text string, trim bool) string {
	if trim {
		return strings.TrimSpace(text)
	}
	return text
}

func lastSegment(path, separator string) string {
	segments := strings.Split(path, separator)
	return segments[len(segments)-1]
}

func commentLines(notices []string, marker string) string {
	lines := ""
	for _, notice := range notices {
		lines += marker + " " + notice + "\n"
	}
	return lines
}
