package logger

// Redactable values are logged through LogValue so secrets never reach the log.
type Redactable interface {
	LogValue() any
}

func redact(metadata []any) []any {
	if len(metadata) == 0 {
		return nil
	}
	result := make([]any, 0, len(metadata))
	for _, item := range metadata {
		result = append(result, redactValue(item))
	}
	return result
}

func redactValue(value any) any {
	switch typed := value.(type) {
	case Redactable:
		return typed.LogValue()
	case map[string]any:
		redacted := make(map[string]any, len(typed))
		for key, item := range typed {
			redacted[key] = redactValue(item)
		}
		return redacted
	default:
		return value
	}
}
