package mcp

import (
	"context"
	"path/filepath"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/fileutils"
)

const responseBodyFileMode = 0o644

type savedResponseBody struct {
	Path string `json:"path"`
}

type saveResponseBodyArguments struct {
	Path string `json:"path"`
}

// saveResponseBody writes this session's last response body as AI output: text masked, binary only when no known secret form appears in it.
func (t *tools) saveResponseBody(ctx context.Context, call Call) (Result, error) {
	var arguments saveResponseBodyArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	if !filepath.IsAbs(arguments.Path) {
		return Result{}, invalidArguments(call.Tool, "path must be an absolute file path")
	}
	path := filepath.Clean(arguments.Path)
	target := dto.AIActivityTarget{Path: path}
	result, err := t.controllers.RequestResponse.LastResult(ctx, dto.NewTabReference(call.TabId()))
	if err != nil {
		return Result{Target: target}, err
	}
	if isBinaryResponse(*result) {
		err = t.saveBinaryResponseBody(ctx, call, path)
	} else {
		err = t.saveTextResponseBody(ctx, call, *result, path)
	}
	if err != nil {
		return Result{Target: target}, err
	}
	return Result{Payload: savedResponseBody{Path: path}, Target: target}, nil
}

func (t *tools) saveTextResponseBody(ctx context.Context, call Call, result dto.SendResult, path string) error {
	text, err := t.responseText(ctx, call, result)
	if err != nil {
		return err
	}
	if err := fileutils.WriteFileAtomic(path, []byte(call.Masked(ctx, text)), responseBodyFileMode); err != nil {
		return invalidArguments(call.Tool, "the body could not be written to "+path)
	}
	return nil
}

func (t *tools) saveBinaryResponseBody(ctx context.Context, call Call, path string) error {
	raw, err := t.responseBytes(ctx, call)
	if err != nil {
		return err
	}
	if call.Masked(ctx, string(raw)) != string(raw) {
		return responseBodySecretRefused()
	}
	return t.controllers.RequestResponse.SaveResponseBody(ctx, dto.SaveResponseBody{TabId: call.TabId(), Path: path})
}
