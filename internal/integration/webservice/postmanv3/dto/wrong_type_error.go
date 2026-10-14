package dto

import "strconv"

type WrongTypeError struct {
	Line int
}

func NewWrongTypeError(line int) WrongTypeError {
	return WrongTypeError{Line: line}
}

func (w WrongTypeError) Error() string {
	return "line " + strconv.Itoa(w.Line) + ": a value has the wrong type"
}
