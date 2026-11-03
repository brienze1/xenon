package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

func TestRetryMCPEditStopsOnSuccess(t *testing.T) {
	attempts := 0
	result, err := retryMCPEdit(context.Background(), func() (string, error) {
		attempts++
		if attempts < 3 {
			return "", errs.ConflictError("the request changed on disk since it was opened", "REQ-02409")
		}
		return "saved", nil
	})
	if err != nil || result != "saved" || attempts != 3 {
		t.Fatalf("retryMCPEdit() = %q, %v after %d attempts", result, err, attempts)
	}
}

func TestRetryMCPEditRetriesEnvironmentConflicts(t *testing.T) {
	attempts := 0
	_, err := retryMCPEdit(context.Background(), func() (int, error) {
		attempts++
		if attempts == 1 {
			return 0, errs.ConflictError("the environment changed on disk since it was opened", "ENV-19409")
		}
		return attempts, nil
	})
	if err != nil || attempts != 2 {
		t.Fatalf("retryMCPEdit() error = %v after %d attempts", err, attempts)
	}
}

func TestRetryMCPEditDoesNotRetryOtherErrors(t *testing.T) {
	failures := []error{
		errs.ConflictError("an environment with this name already exists", "ENV-02409"),
		errs.NotFoundError("the request does not exist", "REQ-01404"),
		errors.New("plain failure"),
	}
	for _, failure := range failures {
		attempts := 0
		_, err := retryMCPEdit(context.Background(), func() (string, error) {
			attempts++
			return "", failure
		})
		if !errors.Is(err, failure) || attempts != 1 {
			t.Errorf("retryMCPEdit() with %v = %v after %d attempts", failure, err, attempts)
		}
	}
}

func TestRetryMCPEditGivesUpAfterTheCap(t *testing.T) {
	attempts := 0
	_, err := retryMCPEdit(context.Background(), func() (string, error) {
		attempts++
		return "", errs.ConflictError("the request changed on disk since it was opened", "REQ-02409")
	})
	if attempts != mcpEditMaxAttempts || !errs.IsConflictError(err) {
		t.Fatalf("retryMCPEdit() = %v after %d attempts, want a conflict after %d", err, attempts, mcpEditMaxAttempts)
	}
}
