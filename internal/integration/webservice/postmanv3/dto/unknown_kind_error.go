package dto

type UnknownKindError struct {
	Kind string
}

func (u UnknownKindError) Error() string {
	return "unknown $kind \"" + u.Kind + "\""
}
