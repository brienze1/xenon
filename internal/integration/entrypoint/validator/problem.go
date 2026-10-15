package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

const (
	invalidProblem     = "is not valid"
	oneOfJoiner        = ", "
	charactersSuffix   = " characters"
	itemsSuffix        = " items"
	atMostLength       = "must be at most "
	atMostItems        = "must have at most "
	atLeastLength      = "must be at least "
	atLeastItems       = "must have at least "
	greaterThan        = "must be greater than "
	atLeastValue       = "must be at least "
	oneOfPrefix        = "must be one of "
	requiredRule       = "required"
	requiredIfRule     = "required_if"
	requiredUnlessRule = "required_unless"
	notBlankRule       = "not_blank"
	oneOfRule          = "oneof"
	maxRule            = "max"
	minRule            = "min"
	gtRule             = "gt"
	gteRule            = "gte"
	emailRule          = "email"
	absolutePathRule   = "absolute_path"
	hasIdRule          = "has_id"
	variableScopeRule  = "variable_scope"
	importInputRule    = "import_input"
	safeSubfolderRule  = "safe_subfolder"
)

var fixedProblems = map[string]string{
	requiredRule:       "is required",
	requiredIfRule:     "is required here",
	requiredUnlessRule: "is required here",
	notBlankRule:       "must not be blank",
	emailRule:          "must be an email address",
	absolutePathRule:   "must be an absolute path",
	hasIdRule:          "must have an id",
	variableScopeRule:  "must be a variable scope (globals, env:<id> or collection:<id>)",
	importInputRule:    "needs at least one path or some pasted text",
	safeSubfolderRule:  "must be a folder inside the workspace",
}

func problemOf(fieldError validator.FieldError) string {
	if problem, found := fixedProblems[fieldError.Tag()]; found {
		return problem
	}
	param := fieldError.Param()
	switch fieldError.Tag() {
	case oneOfRule:
		return oneOfPrefix + strings.Join(strings.Fields(param), oneOfJoiner)
	case maxRule:
		return sizedProblem(fieldError.Kind(), atMostLength, atMostItems, param)
	case minRule:
		return sizedProblem(fieldError.Kind(), atLeastLength, atLeastItems, param)
	case gtRule:
		return greaterThan + param
	case gteRule:
		return atLeastValue + param
	default:
		return invalidProblem
	}
}

func sizedProblem(kind reflect.Kind, lengthPrefix, itemsPrefix, param string) string {
	switch kind {
	case reflect.Slice, reflect.Array, reflect.Map:
		return itemsPrefix + param + itemsSuffix
	case reflect.String:
		return lengthPrefix + param + charactersSuffix
	default:
		return lengthPrefix + param
	}
}
