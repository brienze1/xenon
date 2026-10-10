package steps

import (
	"bytes"
	"fmt"
	"regexp"
)

func expectSameBytes(actual, expected []byte, what string) error {
	if !bytes.Equal(actual, expected) {
		return fmt.Errorf("%s: expected %d bytes, got %d different bytes", what, len(expected), len(actual))
	}
	return nil
}

func expectMatch(actual, pattern, what string) error {
	expression, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern %q: %w", pattern, err)
	}
	if !expression.MatchString(actual) {
		return fmt.Errorf("%s: %q does not match %q", what, actual, pattern)
	}
	return nil
}

func expectDistinct(values []string, what string) error {
	seen := make(map[string]int, len(values))
	for position, value := range values {
		if previous, repeated := seen[value]; repeated {
			return fmt.Errorf("%s %d and %d are both %q", what, previous+1, position+1, value)
		}
		seen[value] = position
	}
	return nil
}
