package mcp

import (
	"context"
	"errors"
	"io"
	"math"
	"mime"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/approvaloutcome"
	"github.com/brienze1/xenon/internal/domain/enums/teststatus"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/errs"
)

const (
	unconfirmedSendCode    = "SND-05409"
	environmentIdAttribute = "environment_id"
	environmentAttribute   = "name"
	colorAttribute         = "color"
)

type sendArguments struct {
	Request          *dto.Request `json:"request"`
	RequestId        string       `json:"request_id"`
	SendUnresolved   bool         `json:"send_unresolved"`
	SendExpiredToken bool         `json:"send_expired_token"`
}

type sendResponseView struct {
	Error       *dto.SendResultError `json:"error,omitempty"`
	StatusText  string               `json:"status_text,omitempty"`
	Body        string               `json:"body,omitempty"`
	ContentType string               `json:"content_type,omitempty"`
	Headers     [][]string           `json:"headers,omitempty"`
	Tests       []sendTestLine       `json:"tests,omitempty"`
	Status      int                  `json:"status,omitempty"`
	Ms          int64                `json:"ms"`
	Size        int64                `json:"size,omitempty"`
	Truncated   bool                 `json:"truncated,omitempty"`
	Binary      bool                 `json:"binary,omitempty"`
}

type sendTestLine struct {
	Name   string `json:"name"`
	Error  string `json:"error,omitempty"`
	Passed bool   `json:"passed"`
}

func (t *tools) send(ctx context.Context, call Call) (Result, error) {
	var arguments sendArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	request, target, err := t.sendTarget(ctx, call, arguments)
	if err != nil {
		return Result{Target: target}, err
	}
	send := dto.SendRequest{
		Origin:                    dto.AISendOrigin(call.Client),
		LocalId:                   call.LocalId,
		TabId:                     call.TabId(),
		Request:                   request,
		AcknowledgeMissingSecrets: arguments.SendUnresolved,
		SendWithExpiredToken:      arguments.SendExpiredToken,
	}
	result, err := t.controllers.RequestSend.Send(ctx, send)
	if failedWithSendCode(err, unconfirmedSendCode) {
		result, err = t.sendConfirmed(ctx, call, send, err)
	}
	if err != nil {
		return Result{Target: target}, err
	}
	view, err := t.sendResponseView(ctx, call, *result)
	changes := []dto.AIChange{{Kind: aichangekind.History.String()}, {Kind: aichangekind.Cookies.String()}}
	return Result{Payload: view, Target: target, Changes: changes, Recorded: true}, err
}

func (t *tools) sendTarget(ctx context.Context, call Call, arguments sendArguments) (dto.Request, dto.AIActivityTarget, error) {
	if arguments.RequestId != "" && arguments.Request != nil {
		return dto.Request{}, dto.AIActivityTarget{}, invalidArguments(call.Tool, "pass request_id or request, not both")
	}
	if arguments.Request != nil {
		return *arguments.Request, dto.AIActivityTarget{Name: arguments.Request.Method + " " + arguments.Request.Url}, nil
	}
	if arguments.RequestId == "" {
		return dto.Request{}, dto.AIActivityTarget{}, invalidArguments(call.Tool, "request_id or request is required")
	}
	index, err := t.treeIndex(ctx, call.LocalId)
	if err != nil {
		return dto.Request{}, dto.AIActivityTarget{}, err
	}
	entry, found := index.entries[arguments.RequestId]
	if !found || !entry.isRequest() {
		return dto.Request{}, dto.AIActivityTarget{Id: arguments.RequestId}, itemNotFound(arguments.RequestId)
	}
	saved, err := t.controllers.RequestGet.Get(ctx, dto.OpenRequest{LocalId: call.LocalId, Id: arguments.RequestId})
	if err != nil {
		return dto.Request{}, entry.target(), err
	}
	return *saved, entry.target(), nil
}

// sendConfirmed resends once with the environment confirmed: at once when approvals are skipped, otherwise after the human approves in Xenon.
func (t *tools) sendConfirmed(ctx context.Context, call Call, send dto.SendRequest, unconfirmed error) (*dto.SendResult, error) {
	environment := unconfirmedEnvironment(unconfirmed)
	settings, err := t.controllers.Settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	if !skipsApprovals(*settings) {
		approval, err := t.controllers.AIAccess.RequestApproval(ctx, dto.AIApprovalRequest{SessionId: call.Session, Client: call.Client, Method: send.Request.Method, Url: send.Request.Url, Environment: environment})
		if err != nil {
			return nil, err
		}
		if err := approvalRefusal(*approval); err != nil {
			return nil, err
		}
	}
	send.ConfirmedEnvironmentId = environment.Id
	return t.controllers.RequestSend.Send(ctx, send)
}

func (t *tools) sendResponseView(ctx context.Context, call Call, result dto.SendResult) (sendResponseView, error) {
	view := sendResponseView{
		Status:      result.StatusCode,
		StatusText:  result.StatusText,
		Ms:          int64(math.Round(result.Timings.TotalMs)),
		ContentType: result.ContentType,
		Error:       result.Error,
		Size:        result.SizeBytes,
		Binary:      isBinaryResponse(result),
		Headers:     make([][]string, 0, len(result.Headers)),
		Tests:       make([]sendTestLine, 0, len(result.Tests)),
	}
	for _, header := range result.Headers {
		view.Headers = append(view.Headers, []string{header.Key, header.Value})
	}
	for _, test := range result.Tests {
		view.Tests = append(view.Tests, sendTestLine{Name: test.Name, Passed: test.Status == teststatus.Passed.String(), Error: test.ErrorMessage})
	}
	if !result.Ok || view.Binary {
		return view, nil
	}
	text, err := t.responseText(ctx, call, result)
	if err != nil {
		return view, err
	}
	body, truncated, size := InlineBody(call.Masked(ctx, text))
	view.Body, view.Truncated, view.Size = body, truncated, int64(size)
	return view, nil
}

func (t *tools) responseText(ctx context.Context, call Call, result dto.SendResult) (string, error) {
	if result.InlineBody != nil {
		return *result.InlineBody, nil
	}
	raw, err := t.responseBytes(ctx, call)
	return string(raw), err
}

func (t *tools) responseBytes(ctx context.Context, call Call) ([]byte, error) {
	body, err := t.controllers.RequestResponse.OpenBody(ctx, dto.NewTabReference(call.TabId()))
	if err != nil {
		return nil, err
	}
	defer body.Content.Close()
	return io.ReadAll(body.Content)
}

// isBinaryResponse also treats a body declared application/octet-stream as binary, so the AI never gets it inlined.
func isBinaryResponse(result dto.SendResult) bool {
	mediaType, _, _ := mime.ParseMediaType(result.ContentType)
	return result.IsBinary || mediaType == entity.MediaTypeOctetStream
}

func unconfirmedEnvironment(err error) dto.AIApprovalEnvironment {
	var custom errs.BaseError
	if !errors.As(err, &custom) {
		return dto.AIApprovalEnvironment{}
	}
	environment := dto.AIApprovalEnvironment{}
	for _, detail := range custom.ToDto("").Error.ErrorDetails {
		if len(detail.Messages) == 0 {
			continue
		}
		switch detail.Attribute {
		case environmentIdAttribute:
			environment.Id = detail.Messages[0]
		case environmentAttribute:
			environment.Name = detail.Messages[0]
		case colorAttribute:
			environment.Color = detail.Messages[0]
		}
	}
	return environment
}

func approvalRefusal(approval dto.AIApprovalOutcome) error {
	switch approval.Outcome {
	case approvaloutcome.Approved.String():
		return nil
	case approvaloutcome.TimedOut.String():
		return approvalTimedOut()
	default:
		return sendDenied()
	}
}

func skipsApprovals(settings dto.AppSettings) bool {
	return settings.AIAccess != nil && settings.AIAccess.SkipApprovals != nil && *settings.AIAccess.SkipApprovals
}
