package entity

import (
	"errors"
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/updatefailure"
	"github.com/brienze1/xenon/pkg/errs"
)

func TestUpdateFailureExposesTheCodeOfItsCause(t *testing.T) {
	cause := errs.ServiceUnavailableError(errors.New("dial tcp: timeout"), "TST-00503", "The server could not be reached")

	failure := NewUpdateFailure(updatefailure.DownloadFailed, cause)

	if failure.Code() != "TST-00503" || failure.Description() != "The server could not be reached" || failure.StatusCode() != 503 {
		t.Fatalf("got %q, %q, %d; want the cause's code, description and status", failure.Code(), failure.Description(), failure.StatusCode())
	}
	if !errors.Is(failure, cause) || failure.Error() != "download_failed: service unavailable" {
		t.Fatalf("the failure must keep wrapping its cause, got %q", failure.Error())
	}
}

func TestUpdateFailureWithoutACodedCauseFallsBackToItsReason(t *testing.T) {
	for name, cause := range map[string]error{"no cause": nil, "plain cause": errors.New("boom")} {
		t.Run(name, func(t *testing.T) {
			failure := NewUpdateFailure(updatefailure.Unverified, cause)

			if failure.Code() != "" || failure.Description() != string(updatefailure.Unverified) || failure.StatusCode() != 0 {
				t.Fatalf("got %q, %q, %d; want the reason only", failure.Code(), failure.Description(), failure.StatusCode())
			}
		})
	}
}
