package dto

import "strconv"

type WrongTypeError struct {
	Line int
}

func (w WrongTypeError) Error() string {
	return "line " + strconv.Itoa(w.Line) + ": a value has the wrong type"
}
