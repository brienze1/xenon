package fileutils

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	maxSafeNameLength = 64
	untitledName      = "untitled"
	trimmedEdges      = "-. "
)

var windowsReservedNames = map[string]bool{
	"con": true, "prn": true, "aux": true, "nul": true,
	"com1": true, "com2": true, "com3": true, "com4": true, "com5": true, "com6": true, "com7": true, "com8": true, "com9": true,
	"lpt1": true, "lpt2": true, "lpt3": true, "lpt4": true, "lpt5": true, "lpt6": true, "lpt7": true, "lpt8": true, "lpt9": true,
}

// SafeName turns a display name into a file name valid on macOS, Windows and Linux (lower-case, at most 64 runes).
func SafeName(name string) string {
	safe := strings.Trim(collapseDashes(replaceUnsafe(strings.ToLower(norm.NFC.String(name)))), trimmedEdges)
	safe = strings.Trim(truncateRunes(safe, maxSafeNameLength), trimmedEdges)
	if safe == "" {
		return untitledName
	}
	return escapeReserved(safe)
}

func replaceUnsafe(name string) string {
	var builder strings.Builder
	for _, char := range name {
		if isUnsafe(char) {
			builder.WriteRune('-')
			continue
		}
		builder.WriteRune(char)
	}
	return builder.String()
}

func isUnsafe(char rune) bool {
	return strings.ContainsRune(`<>:"/\|?*`, char) || unicode.IsControl(char) || unicode.IsSpace(char)
}

func collapseDashes(name string) string {
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}
	return name
}

func truncateRunes(name string, limit int) string {
	runes := []rune(name)
	if len(runes) <= limit {
		return name
	}
	return string(runes[:limit])
}

func escapeReserved(name string) string {
	stem, rest, hasDot := strings.Cut(name, ".")
	if !windowsReservedNames[stem] {
		return name
	}
	if hasDot {
		return stem + "_." + rest
	}
	return stem + "_"
}
