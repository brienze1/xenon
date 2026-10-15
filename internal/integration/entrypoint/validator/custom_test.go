package validator

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

type sample struct {
	Name   string `json:"name" validate:"required,not_blank,max=5"`
	Path   string `json:"path" validate:"absolute_path"`
	Hidden string `json:"-"`
}

func TestStructAcceptsValidInput(t *testing.T) {
	if err := CustomValidator().Struct("TST-00400", &sample{Name: "Acme", Path: absolute()}); err != nil {
		t.Fatalf("Struct() = %v, want nil", err)
	}
}

func TestStructRejectsInvalidFieldsWithTheirJsonNames(t *testing.T) {
	cases := []struct {
		name      string
		input     sample
		attribute string
		problem   string
	}{
		{name: "blank name", input: sample{Name: "   ", Path: absolute()}, attribute: "name", problem: "must not be blank"},
		{name: "long name", input: sample{Name: "abcdef", Path: absolute()}, attribute: "name", problem: "must be at most 5 characters"},
		{name: "relative path", input: sample{Name: "Acme", Path: "relative/dir"}, attribute: "path", problem: "must be an absolute path"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := CustomValidator().Struct("TST-00400", &tc.input)
			var custom errs.BaseError
			if !errors.As(err, &custom) {
				t.Fatalf("Struct() = %v, want BaseError", err)
			}
			dto := custom.ToDto("tx")
			if dto.Error.Code != "TST-00400" || len(dto.Error.ErrorDetails) != 1 {
				t.Fatalf("unexpected error %+v", dto)
			}
			detail := dto.Error.ErrorDetails[0]
			if detail.Attribute != tc.attribute || len(detail.Messages) != 1 || detail.Messages[0] != tc.problem {
				t.Fatalf("detail = %+v, want %s %s", detail, tc.attribute, tc.problem)
			}
			if dto.Error.Description != "invalid request: "+tc.attribute+" "+tc.problem {
				t.Fatalf("description = %q", dto.Error.Description)
			}
		})
	}
}

type ruleRow struct {
	Name string `json:"name" validate:"required"`
}

type rules struct {
	Required       string    `json:"required" validate:"required"`
	Mode           string    `json:"mode"`
	RequiredIf     string    `json:"required_if" validate:"required_if=Mode on"`
	RequiredUnless string    `json:"required_unless" validate:"required_unless=Mode off"`
	NotBlank       string    `json:"not_blank" validate:"not_blank"`
	OneOf          string    `json:"one_of" validate:"oneof=replace keep_both skip"`
	MaxString      string    `json:"max_string" validate:"max=3"`
	MaxSlice       []string  `json:"max_slice" validate:"max=1"`
	MinString      string    `json:"min_string" validate:"min=2"`
	MinSlice       []string  `json:"min_slice" validate:"min=2"`
	Gt             int       `json:"gt" validate:"gt=0"`
	Gte            int       `json:"gte" validate:"gte=1"`
	Email          string    `json:"email" validate:"email"`
	AbsolutePath   string    `json:"absolute_path" validate:"absolute_path"`
	Scope          string    `json:"scope" validate:"variable_scope"`
	Text           string    `json:"text" validate:"import_input"`
	Paths          []string  `json:"paths"`
	Subfolder      string    `json:"subfolder" validate:"safe_subfolder"`
	Unknown        string    `json:"unknown" validate:"uuid"`
	Rows           []ruleRow `json:"rows" validate:"dive"`
}

func TestStructDescribesEveryRuleInPlainEnglish(t *testing.T) {
	input := rules{
		Mode: "on", OneOf: "merge", MaxString: "abcd", MaxSlice: []string{"a", "b"}, MinString: "a", MinSlice: []string{"a"},
		Gt: 0, Gte: 0, Email: "not-an-email", AbsolutePath: "relative", Scope: "nowhere", Subfolder: "../outside", Unknown: "x",
		Rows: []ruleRow{{Name: "ok"}, {Name: ""}},
	}
	err := CustomValidator().Struct("TST-00400", &input)
	var custom errs.BaseError
	if !errors.As(err, &custom) {
		t.Fatalf("Struct() = %v, want BaseError", err)
	}
	problems := map[string]string{}
	for _, detail := range custom.ToDto("tx").Error.ErrorDetails {
		problems[detail.Attribute] = detail.Messages[0]
	}
	want := map[string]string{
		"required":      "is required",
		"required_if":   "is required here",
		"not_blank":     "must not be blank",
		"one_of":        "must be one of replace, keep_both, skip",
		"max_string":    "must be at most 3 characters",
		"max_slice":     "must have at most 1 items",
		"min_string":    "must be at least 2 characters",
		"min_slice":     "must have at least 2 items",
		"gt":            "must be greater than 0",
		"gte":           "must be at least 1",
		"email":         "must be an email address",
		"absolute_path": "must be an absolute path",
		"scope":         "must be a variable scope (globals, env:<id> or collection:<id>)",
		"text":          "needs at least one path or some pasted text",
		"subfolder":     "must be a folder inside the workspace",
		"unknown":       "is not valid",
		"rows[1].name":  "is required",
	}
	for attribute, problem := range want {
		if problems[attribute] != problem {
			t.Errorf("%s = %q, want %q", attribute, problems[attribute], problem)
		}
	}
	unless := rules{Mode: "on"}
	err = CustomValidator().Struct("TST-00400", &unless)
	errors.As(err, &custom)
	for _, detail := range custom.ToDto("tx").Error.ErrorDetails {
		if detail.Attribute == "required_unless" && detail.Messages[0] != "is required here" {
			t.Errorf("required_unless = %q", detail.Messages[0])
		}
	}
}

func TestStructJoinsEveryFailingFieldIntoTheDescription(t *testing.T) {
	err := CustomValidator().Struct("TST-00400", &sample{Name: " ", Path: "relative"})
	var custom errs.BaseError
	if !errors.As(err, &custom) {
		t.Fatalf("Struct() = %v", err)
	}
	if got := custom.Description(); got != "invalid request: name must not be blank; path must be an absolute path" {
		t.Errorf("Description() = %q", got)
	}
}

func TestHasIdProblemWording(t *testing.T) {
	type identified struct {
		Id string
	}
	type input struct {
		Identified identified `json:"identified" validate:"has_id"`
	}
	err := CustomValidator().Struct("TST-00400", &input{})
	var custom errs.BaseError
	if !errors.As(err, &custom) || custom.ToDto("tx").Error.ErrorDetails[0].Messages[0] != "must have an id" {
		t.Errorf("has_id problem = %v", err)
	}
}

func TestStructRejectsNonStructInput(t *testing.T) {
	err := CustomValidator().Struct("TST-01400", "not a struct")
	var custom errs.BaseError
	if !errors.As(err, &custom) || custom.Code() != "TST-01400" {
		t.Fatalf("Struct() = %v, want TST-01400", err)
	}
}

func absolute() string {
	path, _ := filepath.Abs("workspace")
	return path
}

func TestHasIdRequiresANonBlankIdField(t *testing.T) {
	type identified struct {
		Id string
	}
	type anonymous struct {
		Name string
	}
	type input struct {
		Identified identified `json:"identified" validate:"has_id"`
	}
	type withoutId struct {
		Anonymous anonymous `json:"anonymous" validate:"has_id"`
	}
	if err := CustomValidator().Struct("TST-00400", &input{Identified: identified{Id: "x"}}); err != nil {
		t.Fatalf("an id is valid: %v", err)
	}
	for name, value := range map[string]any{"blank": &input{Identified: identified{Id: "  "}}, "missing field": &withoutId{}} {
		if err := CustomValidator().Struct("TST-00400", value); err == nil {
			t.Errorf("%s must be rejected", name)
		}
	}
}
