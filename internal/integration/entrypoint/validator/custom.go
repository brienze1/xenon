package validator

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/errs"
)

type Validate interface {
	Struct(code string, s any) error
}

type customValidator struct {
	validator *validator.Validate
}

func CustomValidator() Validate {
	validate := validator.New()
	_ = validate.RegisterValidation("not_blank", notBlank)
	_ = validate.RegisterValidation("absolute_path", absolutePath)
	_ = validate.RegisterValidation("has_id", hasId)
	_ = validate.RegisterValidation("variable_scope", variableScope)
	_ = validate.RegisterValidation("import_input", importInput)
	validate.RegisterTagNameFunc(jsonName)
	return &customValidator{validator: validate}
}

func (c *customValidator) Struct(code string, s any) error {
	err := c.validator.Struct(s)
	if err == nil {
		return nil
	}
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return errs.BadRequestError(err.Error(), code)
	}
	details := make([]errs.ErrorDetails, 0, len(validationErrors))
	for _, fieldError := range validationErrors {
		details = append(details, errs.ErrorDetails{
			Attribute: fieldError.Field(),
			Messages:  []string{"failed on the '" + fieldError.Tag() + "' rule"},
		})
	}
	return errs.BadRequestError("invalid request", code, details...)
}

func notBlank(field validator.FieldLevel) bool {
	return strings.TrimSpace(field.Field().String()) != ""
}

func absolutePath(field validator.FieldLevel) bool {
	return filepath.IsAbs(field.Field().String())
}

func hasId(field validator.FieldLevel) bool {
	id := field.Field().FieldByName("Id")
	return id.IsValid() && strings.TrimSpace(id.String()) != ""
}

func variableScope(field validator.FieldLevel) bool {
	return entity.ParseScopeRef(field.Field().String()).IsValid()
}

func importInput(field validator.FieldLevel) bool {
	paths, _ := field.Parent().FieldByName("Paths").Interface().([]string)
	return entity.ImportRequest{Text: field.Field().String(), Paths: paths}.HasInput()
}

func jsonName(field reflect.StructField) string {
	name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
	if name == "-" {
		return ""
	}
	return name
}
