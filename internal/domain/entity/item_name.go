package entity

import (
	"strings"
	"unicode/utf8"
)

const maxItemNameLength = 200

func NormalizeItemName(name string) string {
	return strings.TrimSpace(name)
}

func isValidItemName(name string) bool {
	length := utf8.RuneCountInString(strings.TrimSpace(name))
	return length > 0 && length <= maxItemNameLength
}
