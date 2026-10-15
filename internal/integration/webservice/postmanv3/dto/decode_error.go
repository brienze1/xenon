package dto

import (
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

const (
	yamlErrorPrefix = "yaml: "
	wrongTypeReason = "a value has the wrong type"
)

var yamlLinePattern = regexp.MustCompile(`^line (\d+):`)

type DecodeError struct {
	Path   string
	Reason string
}

func NewDecodeError(path string, err error) DecodeError {
	return DecodeError{Path: path, Reason: decodeReason(err)}
}

func NewWrongTypeDecodeError(path string) DecodeError {
	return DecodeError{Path: path, Reason: wrongTypeReason}
}

func (d DecodeError) Error() string {
	return d.Path + " is not a valid Postman v3 file: " + d.Reason
}

func decodeReason(err error) string {
	var wrongType WrongTypeError
	var unknownKind UnknownKindError
	var typeError *yaml.TypeError
	var syntaxError *json.SyntaxError
	switch {
	case errors.As(err, &wrongType):
		return wrongType.Error()
	case errors.As(err, &unknownKind):
		return unknownKind.Error()
	case errors.As(err, &syntaxError):
		return "byte " + strconv.FormatInt(syntaxError.Offset, 10) + ": invalid JSON"
	case errors.As(err, &typeError):
		return typeErrorReason(typeError)
	default:
		return strings.TrimPrefix(err.Error(), yamlErrorPrefix)
	}
}

func typeErrorReason(typeError *yaml.TypeError) string {
	if len(typeError.Errors) > 0 {
		if match := yamlLinePattern.FindStringSubmatch(typeError.Errors[0]); match != nil {
			return "line " + match[1] + ": " + wrongTypeReason
		}
	}
	return wrongTypeReason
}
