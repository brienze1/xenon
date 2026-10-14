package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/historysource"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

const secretToken = "super-secret-token-123"

type logBuffer struct {
	buffer bytes.Buffer
	mu     sync.Mutex
}

func (l *logBuffer) Write(data []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buffer.Write(data)
}

func (l *logBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buffer.String()
}

type echoSendService struct{}

func (echoSendService) Send(_ context.Context, _, _ string, request entity.Request, _ string, _ entity.SendAcknowledgement, _ entity.SendOrigin) (*entity.SendExecution, error) {
	headers := []entity.KeyValue{entity.NewKeyValue("Set-Cookie", "session="+secretToken)}
	execution := entity.NewSendExecution("send")
	execution.RecordResponse(entity.EffectiveRequest{}, entity.NewResponse(200, "OK", "HTTP/1.1", headers, []byte(request.Auth.Bearer.Token), 10, entity.Timings{}))
	return &execution, nil
}

type echoEditService struct{}

func (echoEditService) ApplyUrl(_ context.Context, request entity.Request) entity.Request {
	return request
}
func (echoEditService) ApplyParams(_ context.Context, request entity.Request) entity.Request {
	return request
}
func (echoEditService) ParseBulk(_ context.Context, text string) []entity.KeyValue {
	return []entity.KeyValue{entity.NewKeyValue("Authorization", text)}
}
func (echoEditService) ParseCurl(_ context.Context, text string) (*entity.ParsedCurl, error) {
	request := entity.NewRequest("", "")
	request.URL = "https://api.example.test/me"
	request.Headers = []entity.KeyValue{entity.NewKeyValue("Authorization", text)}
	return &entity.ParsedCurl{Request: request, Notices: []string{}}, nil
}
func (echoEditService) FormatBulk(_ context.Context, rows []entity.KeyValue) string {
	return rows[0].Value
}
func (echoEditService) AutoHeaders(_ context.Context, _ string, request entity.Request, _ string) ([]entity.AutoHeader, error) {
	return []entity.AutoHeader{entity.NewAutoHeader("Authorization", "Bearer "+request.Auth.Bearer.Token)}, nil
}

type storedResponseService struct{}

func (storedResponseService) LastResult(context.Context, string) (*entity.Response, error) {
	response := entity.NewResponse(200, "OK", "HTTP/1.1", nil, []byte(secretToken), 10, entity.Timings{})
	return &response, nil
}
func (storedResponseService) OpenBody(context.Context, string) (*entity.Response, io.ReadCloser, error) {
	response := entity.NewResponse(200, "OK", "HTTP/1.1", nil, nil, 0, entity.Timings{})
	return &response, io.NopCloser(strings.NewReader(secretToken)), nil
}
func (storedResponseService) SaveBody(context.Context, string, string) error { return nil }
func (storedResponseService) Release(context.Context, string) error          { return nil }
func (storedResponseService) ReleaseAll(context.Context) error               { return nil }

const secretRequestJSON = `{
	"method": "POST",
	"url": "https://api.example.test/me",
	"auth": { "type": "bearer", "bearer": { "token": "` + secretToken + `" } },
	"headers": [ { "key": "X-Api-Key", "value": "` + secretToken + `" } ],
	"body": { "mode": "raw", "raw": { "content": "` + secretToken + `" } }
}`

func decodeDTO[T any](t *testing.T, raw string) T {
	t.Helper()
	var decoded T
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		t.Fatal(err)
	}
	return decoded
}

func captureLogs(t *testing.T) *logBuffer {
	t.Helper()
	t.Setenv("LOG_LEVEL", "trace")
	buffer := &logBuffer{}
	logger.SetOutput(buffer)
	t.Cleanup(func() { logger.SetOutput(os.Stdout) })
	return buffer
}

func TestRequestControllersNeverLogSecrets(t *testing.T) {
	logs := captureLogs(t)
	ctx := context.Background()
	edit := NewRequestEdit(NewErrorHandler(), echoEditService{})
	send := NewRequestSend(NewErrorHandler(), echoSendService{})
	response := NewRequestResponse(NewErrorHandler(), storedResponseService{})

	sent, err := send.Send(ctx, dto.NewSendRequest(dto.NewSendOrigin(historysource.User, ""), "ws", "tab", "", decodeDTO[dto.Request](t, secretRequestJSON), false, false))
	if err != nil || *sent.InlineBody != secretToken {
		t.Fatalf("send = %+v %v", sent, err)
	}
	if _, err := edit.ApplyUrl(ctx, decodeDTO[dto.ApplyRequest](t, `{"request": `+secretRequestJSON+`}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := edit.ApplyParams(ctx, decodeDTO[dto.ApplyRequest](t, `{"request": `+secretRequestJSON+`}`)); err != nil {
		t.Fatal(err)
	}
	if headers, err := edit.AutoHeaders(ctx, decodeDTO[dto.AutoHeadersRequest](t, `{"local_id": "ws", "request": `+secretRequestJSON+`}`)); err != nil || headers[0].Value != "Bearer "+secretToken {
		t.Fatalf("auto headers = %+v %v", headers, err)
	}
	if rows, err := edit.ParseBulk(ctx, dto.NewBulkText(secretToken)); err != nil || rows.Rows[0].Value != secretToken {
		t.Fatalf("parse bulk = %+v %v", rows, err)
	}
	if text, err := edit.FormatBulk(ctx, dto.NewBulkRows([]entity.KeyValue{entity.NewKeyValue("Authorization", secretToken)})); err != nil || text.Text != secretToken {
		t.Fatalf("format bulk = %+v %v", text, err)
	}
	if parsed, err := edit.ParseCurl(ctx, decodeDTO[dto.CurlText](t, `{"text": "`+secretToken+`"}`)); err != nil || parsed.Request.Headers[0].Value != secretToken {
		t.Fatalf("parse curl = %+v %v", parsed, err)
	}
	if last, err := response.LastResult(ctx, dto.NewTabReference("tab")); err != nil || *last.InlineBody != secretToken {
		t.Fatalf("last result = %+v %v", last, err)
	}
	if body, err := response.OpenBody(ctx, dto.NewTabReference("tab")); err != nil || body.Content == nil {
		t.Fatalf("open body = %+v %v", body, err)
	}

	output := logs.String()
	if !strings.Contains(output, "https://api.example.test/me") {
		t.Fatalf("the logs must still identify the request: %s", output)
	}
	if strings.Contains(output, secretToken) {
		t.Fatalf("a secret reached the log output:\n%s", output)
	}
}
