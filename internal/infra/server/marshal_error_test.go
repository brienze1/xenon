package server

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/brienze1/xenon/pkg/errs"
)

type errorPayload struct {
	Error struct {
		Id          string `json:"id"`
		Code        string `json:"code"`
		Description string `json:"description"`
	} `json:"error"`
}

func TestMarshalErrorKeepsTheCodeOfCustomErrors(t *testing.T) {
	payload := decode(t, MarshalError(errs.NotFoundError("workspace folder is missing", "WKS-06404")))
	if payload.Error.Code != "WKS-06404" || payload.Error.Description != "workspace folder is missing" {
		t.Fatalf("payload = %+v", payload)
	}
	if len(payload.Error.Id) != 32 {
		t.Fatalf("id = %q, want 32 hex chars", payload.Error.Id)
	}
}

func TestMarshalErrorWrapsUnknownErrors(t *testing.T) {
	payload := decode(t, MarshalError(errors.New("boom")))
	if payload.Error.Code != "UNK-01500" || payload.Error.Description != "Unknown error" {
		t.Fatalf("payload = %+v", payload)
	}
}

func decode(t *testing.T, raw []byte) errorPayload {
	t.Helper()
	var payload errorPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("invalid json %s: %v", raw, err)
	}
	return payload
}
