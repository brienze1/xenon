package errs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"testing"
)

type expectation struct {
	err      BaseError
	name     string
	message  string
	code     string
	status   int
	retries  int
	abort    bool
	notified bool
}

func TestFactories(t *testing.T) {
	cases := []expectation{
		{name: "bad request", err: BadRequestError("invalid", "TST-00400"), message: "bad request", code: "TST-00400", status: http.StatusBadRequest, retries: 0, abort: true, notified: true},
		{name: "not found", err: NotFoundError("missing", "TST-00404"), message: "not found", code: "TST-00404", status: http.StatusNotFound, retries: 0, abort: true, notified: true},
		{name: "internal", err: InternalServerError(errors.New("boom"), "TST-00500", "failed"), message: "internal server error", code: "TST-00500", status: http.StatusInternalServerError, retries: 5, abort: false, notified: true},
		{name: "panic", err: PanicError("panicked"), message: "panic", code: "PNC-00500", status: http.StatusInternalServerError, retries: 0, abort: false, notified: true},
		{name: "server", err: ServerError("upstream", "TST-00502", http.StatusBadGateway), message: "server error", code: "TST-00502", status: http.StatusBadGateway, retries: 5, abort: false, notified: true},
		{name: "client", err: ClientError("rejected", "TST-00422", http.StatusUnprocessableEntity), message: "client error", code: "TST-00422", status: http.StatusUnprocessableEntity, retries: 0, abort: true, notified: true},
		{name: "conflict", err: ConflictError("taken", "TST-00409"), message: "conflict", code: "TST-00409", status: http.StatusConflict, retries: 0, abort: true, notified: true},
		{name: "forbidden", err: ForbiddenError("denied", "TST-00403"), message: "forbidden", code: "TST-00403", status: http.StatusForbidden, retries: 0, abort: true, notified: true},
		{name: "unprocessable entity", err: UnprocessableEntityError("unreadable", "TST-01422"), message: "unprocessable entity", code: "TST-01422", status: http.StatusUnprocessableEntity, retries: 0, abort: true, notified: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.err.Error() != tc.message {
				t.Errorf("Error() = %q, want %q", tc.err.Error(), tc.message)
			}
			if tc.err.Code() != tc.code {
				t.Errorf("Code() = %q, want %q", tc.err.Code(), tc.code)
			}
			if tc.err.StatusCode() != tc.status {
				t.Errorf("StatusCode() = %d, want %d", tc.err.StatusCode(), tc.status)
			}
			if tc.err.Retries() != tc.retries {
				t.Errorf("Retries() = %d, want %d", tc.err.Retries(), tc.retries)
			}
			if tc.err.Abort() != tc.abort {
				t.Errorf("Abort() = %v, want %v", tc.err.Abort(), tc.abort)
			}
			if tc.err.InternalNotify() != tc.notified {
				t.Errorf("InternalNotify() = %v, want %v", tc.err.InternalNotify(), tc.notified)
			}
			dto := tc.err.ToDto("tx-1")
			if dto.Error.Id != "tx-1" || dto.Error.Code != tc.code || dto.Error.Description != tc.err.Description() {
				t.Errorf("ToDto() = %+v", dto)
			}
			if tc.err.ErrorDetails() == nil {
				t.Error("ErrorDetails() returned nil")
			}
			tc.err.SetStatusCode(http.StatusTeapot, "changed")
			if tc.err.StatusCode() != http.StatusTeapot || tc.err.Description() != "changed" {
				t.Errorf("SetStatusCode did not apply: %d %q", tc.err.StatusCode(), tc.err.Description())
			}
			if !tc.err.Is(fmt.Errorf("wrapped: %w", tc.err)) {
				t.Error("Is() should match a wrapped error of the same type")
			}
		})
	}
}

func TestIsDoesNotMatchOtherTypes(t *testing.T) {
	if BadRequestError("a", "b").Is(NotFoundError("a", "b")) {
		t.Error("bad request should not match not found")
	}
}

func TestBadRequestDetails(t *testing.T) {
	err := BadRequestError("invalid", "TST-00400", ErrorDetails{Attribute: "name", Messages: []string{"required"}})
	details := err.ErrorDetails()
	if details["attribute"] != "name" {
		t.Errorf("attribute = %v", details["attribute"])
	}
	if got := err.ToDto("").Error.ErrorDetails; len(got) != 1 || got[0].Messages[0] != "required" {
		t.Errorf("dto details = %+v", got)
	}
}

func TestInternalServerErrorWrapsCause(t *testing.T) {
	cause := errors.New("disk on fire")
	err := InternalServerError(cause, "TST-00500", "failed")
	if !errors.Is(err, cause) {
		t.Error("internal server error should unwrap to its cause")
	}
	if err.ErrorDetails()["attribute"] != "UNKNOWN_ERROR" {
		t.Errorf("details = %v", err.ErrorDetails())
	}
	if InternalServerError(nil, "c", "m").ToDto("").Error.ErrorDetails[0].Messages[0] != "" {
		t.Error("nil cause should produce an empty message")
	}
}

func TestIsNotFoundError(t *testing.T) {
	if !IsNotFoundError(fmt.Errorf("x: %w", NotFoundError("m", "c"))) {
		t.Error("expected not found")
	}
	if IsNotFoundError(errors.New("other")) {
		t.Error("plain error is not a not found error")
	}
}

func TestClientErrorParsesBody(t *testing.T) {
	body := io.NopCloser(strings.NewReader(`{"error":{"error_details":[{"attribute":"id","messages":["bad"]}]}}`))
	err := ClientError("rejected", "TST-00400", http.StatusBadRequest, body)
	if got := err.ToDto("").Error.ErrorDetails; len(got) != 1 || got[0].Attribute != "id" {
		t.Errorf("details = %+v", got)
	}
	invalid := ClientError("rejected", "TST-00400", http.StatusBadRequest, io.NopCloser(strings.NewReader("not json")))
	if len(invalid.ToDto("").Error.ErrorDetails) != 0 {
		t.Error("invalid body should produce no details")
	}
	if len(ClientError("r", "c", 400, nil).ToDto("").Error.ErrorDetails) != 0 {
		t.Error("nil body should produce no details")
	}
}

func TestFileError(t *testing.T) {
	cases := []struct {
		err     error
		name    string
		message string
		status  int
	}{
		{name: "not exist", err: fmt.Errorf("open: %w", fs.ErrNotExist), status: http.StatusNotFound, message: "not found"},
		{name: "permission", err: fmt.Errorf("open: %w", fs.ErrPermission), status: http.StatusForbidden, message: "forbidden"},
		{name: "other", err: errors.New("io"), status: http.StatusInternalServerError, message: "internal server error"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := FileError(tc.err, "FIL-00001")
			if got.StatusCode() != tc.status || got.Code() != "FIL-00001" || got.Error() != tc.message {
				t.Errorf("FileError() = %d %q %q", got.StatusCode(), got.Code(), got.Error())
			}
		})
	}
}
