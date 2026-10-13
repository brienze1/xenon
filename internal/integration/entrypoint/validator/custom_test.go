package validator

import (
	"errors"
	"path/filepath"
	"strings"
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
		rule      string
	}{
		{name: "blank name", input: sample{Name: "   ", Path: absolute()}, attribute: "name", rule: "not_blank"},
		{name: "long name", input: sample{Name: "abcdef", Path: absolute()}, attribute: "name", rule: "max"},
		{name: "relative path", input: sample{Name: "Acme", Path: "relative/dir"}, attribute: "path", rule: "absolute_path"},
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
			if detail.Attribute != tc.attribute || !strings.Contains(detail.Messages[0], "'"+tc.rule+"'") {
				t.Fatalf("detail = %+v, want %s/%s", detail, tc.attribute, tc.rule)
			}
		})
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
