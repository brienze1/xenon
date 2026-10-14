package entity

const AIMaxBodyReadBytes = 65536

type AIBodySlice struct {
	Text      string
	Offset    int
	Size      int
	Truncated bool
}

// NewAIBodySlice reads length bytes (16 KB when 0) of a masked body from offset; false means out-of-range arguments.
func NewAIBodySlice(masked string, offset, length int) (AIBodySlice, bool) {
	if offset < 0 || length < 0 || length > AIMaxBodyReadBytes {
		return AIBodySlice{}, false
	}
	if length == 0 {
		length = AIInlineBodyBytes
	}
	start := runeBoundary(masked, min(offset, len(masked)))
	end := runeBoundary(masked, min(start+length, len(masked)))
	return AIBodySlice{Text: masked[start:end], Offset: start, Size: len(masked), Truncated: end < len(masked)}, true
}
