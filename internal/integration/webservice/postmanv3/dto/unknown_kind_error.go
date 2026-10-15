package dto

type UnknownKindError struct {
	Kind string
}

func NewUnknownKindError(kind string) UnknownKindError {
	return UnknownKindError{Kind: kind}
}

func (u UnknownKindError) Error() string {
	return "unknown $kind \"" + u.Kind + "\""
}
