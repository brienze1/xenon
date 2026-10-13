package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type responseBodyArguments struct {
	Offset int `json:"offset"`
	Length int `json:"length"`
}

type bodySliceView struct {
	Text      string `json:"text"`
	Offset    int    `json:"offset"`
	Size      int    `json:"size"`
	Truncated bool   `json:"truncated,omitempty"`
}

type binaryBodyView struct {
	ContentType string `json:"content_type,omitempty"`
	Size        int64  `json:"size"`
	Binary      bool   `json:"binary"`
}

func (t *tools) response(ctx context.Context, call Call) (Result, error) {
	if call.Action == "last" {
		return t.lastResponse(ctx, call)
	}
	return t.responseBody(ctx, call)
}

func (t *tools) lastResponse(ctx context.Context, call Call) (Result, error) {
	result, err := t.controllers.RequestResponse.LastResult(ctx, dto.NewTabReference(call.TabId()))
	if err != nil {
		return Result{}, err
	}
	view, err := t.sendResponseView(ctx, call, *result)
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: view}, nil
}

func (t *tools) responseBody(ctx context.Context, call Call) (Result, error) {
	var arguments responseBodyArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	result, err := t.controllers.RequestResponse.LastResult(ctx, dto.NewTabReference(call.TabId()))
	if err != nil {
		return Result{}, err
	}
	if isBinaryResponse(*result) {
		return Result{Payload: binaryBodyView{Binary: true, ContentType: result.ContentType, Size: result.SizeBytes}}, nil
	}
	text, err := t.responseText(ctx, call, *result)
	if err != nil {
		return Result{}, err
	}
	slice, err := SliceBody(call.Tool, call.Masked(ctx, text), arguments.Offset, min(arguments.Length, MaxBodyReadBytes))
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: bodySliceView{Text: slice.Text, Offset: slice.Offset, Size: slice.Size, Truncated: slice.Truncated}}, nil
}
