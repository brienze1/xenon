package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

func TestErrorHandlerKeepsCustomErrors(t *testing.T) {
	original := errs.NotFoundError("missing", "TST-01404")
	handled := NewErrorHandler().Handle(context.Background(), original, nil)
	if handled != original {
		t.Fatalf("handled = %v, want the original error", handled)
	}
}

func TestErrorHandlerWrapsUnknownErrors(t *testing.T) {
	handled := NewErrorHandler().Handle(context.Background(), errors.New("boom"), "event")
	var custom errs.BaseError
	if !errors.As(handled, &custom) {
		t.Fatalf("handled = %v, want a BaseError", handled)
	}
	if custom.Code() != "UNK-00500" {
		t.Fatalf("code = %q, want UNK-00500", custom.Code())
	}
}

func TestHandlePanicDescribesEveryRecoveredKind(t *testing.T) {
	var nilMap map[string]int
	runtimeErr := func() (recovered any) {
		defer func() { recovered = recover() }()
		nilMap["x"] = 1
		return nil
	}()
	cases := []struct {
		recovered any
		want      string
	}{
		{recovered: "text", want: "recovered (string) panic: text"},
		{recovered: runtimeErr, want: "recovered (runtime.Error) panic: assignment to entry in nil map"},
		{recovered: errors.New("failure"), want: "recovered (error) panic: failure"},
		{recovered: 42, want: "recovered (default) panic: 42"},
	}
	for _, tc := range cases {
		err := handlePanic(context.Background(), nil, tc.recovered)
		var custom errs.BaseError
		if !errors.As(err, &custom) {
			t.Fatalf("handlePanic returned %v", err)
		}
		if custom.Description() != tc.want || custom.Code() != "PNC-00500" {
			t.Errorf("description = %q code = %q, want %q", custom.Description(), custom.Code(), tc.want)
		}
	}
}
