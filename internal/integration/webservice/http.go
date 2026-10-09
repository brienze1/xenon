package webservice

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"io"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"os"
	"strings"

	"github.com/andybalholm/brotli"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/webservice/dto"
	httpClient "github.com/brienze1/xenon/pkg/http"
	"github.com/brienze1/xenon/pkg/logger"
)

type httpWebservice struct {
	newClient httpClient.ClientFactory
}

func NewHttpWebservice(newClient httpClient.ClientFactory) adapter.HttpWebservice {
	return &httpWebservice{newClient: newClient}
}

func (h *httpWebservice) Send(ctx context.Context, request entity.EffectiveRequest) entity.Response {
	logger.Debug(ctx, "Started", request)

	trace := newTimingTrace()
	target, sendErr := parseTarget(request.URL)
	if sendErr != nil {
		return failed(ctx, *sendErr, trace)
	}
	files, sendErr := readFiles(request.Body.FileRefs())
	if sendErr != nil {
		return failed(ctx, *sendErr, trace)
	}
	body := dto.NewRequestBody(request.Body, files)
	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, target.String(), body.Content)
	if err != nil {
		return failed(ctx, entity.NewSendError(senderrorkind.Other, err.Error()), trace)
	}
	applyHeaders(httpRequest, request, &body)

	client := h.newClient.Build(httpClient.Options{
		Timeout:         request.Settings.Timeout(),
		FollowRedirects: request.Settings.FollowRedirects,
		VerifySsl:       request.Settings.VerifySsl,
	})
	defer client.CloseIdleConnections()
	httpResponse, err := client.Do(httptrace.WithClientTrace(ctx, trace.clientTrace()), httpRequest)
	if err != nil {
		return failed(ctx, sendError(err, request, target), trace)
	}
	defer func() { _ = httpResponse.Body.Close() }()
	raw, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return failed(ctx, sendError(err, request, target), trace)
	}
	response := dto.NewHttpResponse(httpResponse, decode(httpResponse.Header.Get("Content-Encoding"), raw), int64(len(raw))).ToEntity(trace.finish())

	logger.Debug(ctx, "Finished", response)
	return response
}

func parseTarget(raw string) (*url.URL, *entity.SendError) {
	if strings.TrimSpace(raw) == "" {
		sendErr := entity.NewSendError(senderrorkind.InvalidUrl, "")
		return nil, &sendErr
	}
	target, err := url.Parse(raw)
	if err != nil || !isHttpScheme(target.Scheme) || target.Host == "" {
		sendErr := entity.NewSendError(senderrorkind.InvalidUrl, raw)
		return nil, &sendErr
	}
	return target, nil
}

func readFiles(refs []entity.FileRef) (map[string][]byte, *entity.SendError) {
	files := make(map[string][]byte, len(refs))
	for _, ref := range refs {
		content, err := os.ReadFile(ref.Src)
		if os.IsNotExist(err) {
			sendErr := entity.NewSendError(senderrorkind.MissingFile, ref.Src)
			return nil, &sendErr
		}
		if err != nil {
			sendErr := entity.NewSendError(senderrorkind.Other, err.Error())
			return nil, &sendErr
		}
		files[ref.Src] = content
	}
	return files, nil
}

func isHttpScheme(scheme string) bool {
	return strings.EqualFold(scheme, "http") || strings.EqualFold(scheme, "https")
}

func applyHeaders(httpRequest *http.Request, request entity.EffectiveRequest, body *dto.RequestBody) {
	for _, header := range request.SendHeaders() {
		if strings.EqualFold(header.Key, entity.HeaderHost) {
			httpRequest.Host = header.Value
			continue
		}
		httpRequest.Header.Add(header.Key, header.Value)
	}
	if _, found := httpRequest.Header[entity.HeaderUserAgent]; !found {
		httpRequest.Header[entity.HeaderUserAgent] = []string{""}
	}
	if body.ContentType != "" && request.UsesAutoHeader(entity.HeaderContentType) {
		httpRequest.Header.Set(entity.HeaderContentType, body.ContentType)
	}
	if request.IsAutoHeaderDisabled(entity.HeaderContentLength) {
		httpRequest.ContentLength = -1
	}
}

func decode(encoding string, raw []byte) []byte {
	var reader io.Reader
	var err error
	switch strings.ToLower(strings.TrimSpace(encoding)) {
	case "gzip":
		reader, err = gzip.NewReader(bytes.NewReader(raw))
	case "deflate":
		reader, err = zlib.NewReader(bytes.NewReader(raw))
		if err != nil {
			reader, err = flate.NewReader(bytes.NewReader(raw)), nil
		}
	case "br":
		reader = brotli.NewReader(bytes.NewReader(raw))
	default:
		return raw
	}
	if err != nil {
		return raw
	}
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return raw
	}
	return decoded
}

func failed(ctx context.Context, sendErr entity.SendError, trace *timingTrace) entity.Response {
	response := entity.NewFailedResponse(sendErr, trace.finish())
	logger.Debug(ctx, "Finished", response, map[string]any{"kind": sendErr.Kind})
	return response
}
