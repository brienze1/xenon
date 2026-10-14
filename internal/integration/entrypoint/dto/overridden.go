package dto

func overridden[T any](override *T, stored T) T {
	if override == nil {
		return stored
	}
	return *override
}
