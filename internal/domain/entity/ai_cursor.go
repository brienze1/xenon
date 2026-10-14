package entity

import (
	"encoding/base64"
	"strconv"
	"strings"
)

const aiCursorPrefix = "o:"

func encodeAICursor(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(aiCursorPrefix + strconv.Itoa(offset)))
}

// ParseAICursor reads an opaque page cursor; empty is offset 0 and false means an unknown cursor.
func ParseAICursor(raw string) (int, bool) {
	if raw == "" {
		return 0, true
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || !strings.HasPrefix(string(decoded), aiCursorPrefix) {
		return 0, false
	}
	offset, err := strconv.Atoi(strings.TrimPrefix(string(decoded), aiCursorPrefix))
	if err != nil || offset < 0 {
		return 0, false
	}
	return offset, true
}
