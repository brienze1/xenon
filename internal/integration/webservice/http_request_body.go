package webservice

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

type requestBody struct {
	content     io.Reader
	contentType string
}

func newRequestBody(body entity.Body) (*requestBody, *entity.SendError) {
	switch body.Mode {
	case bodymode.Raw:
		return &requestBody{content: strings.NewReader(body.Raw.Content)}, nil
	case bodymode.UrlEncoded:
		return &requestBody{content: strings.NewReader(urlEncoded(body.UrlEncoded))}, nil
	case bodymode.GraphQl:
		return &requestBody{content: bytes.NewReader(graphQl(body.GraphQl))}, nil
	case bodymode.Binary:
		return binaryBody(body.Binary.Src)
	case bodymode.FormData:
		return formDataBody(body.FormData)
	default:
		return &requestBody{}, nil
	}
}

func urlEncoded(rows []entity.KeyValue) string {
	pairs := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.IsActive() {
			pairs = append(pairs, url.QueryEscape(row.Key)+"="+url.QueryEscape(row.Value))
		}
	}
	return strings.Join(pairs, "&")
}

func graphQl(body entity.BodyGraphQl) []byte {
	variables := json.RawMessage("{}")
	if trimmed := strings.TrimSpace(body.Variables); json.Valid([]byte(trimmed)) {
		variables = json.RawMessage(trimmed)
	}
	payload, _ := json.Marshal(map[string]any{"query": body.Query, "variables": variables})
	return payload
}

func binaryBody(path string) (*requestBody, *entity.SendError) {
	content, sendErr := readFile(path)
	if sendErr != nil {
		return nil, sendErr
	}
	return &requestBody{content: bytes.NewReader(content)}, nil
}

func formDataBody(rows []entity.FormDataRow) (*requestBody, *entity.SendError) {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	for _, row := range rows {
		if !row.Enabled {
			continue
		}
		if sendErr := writeFormDataRow(writer, row); sendErr != nil {
			return nil, sendErr
		}
	}
	_ = writer.Close()
	return &requestBody{content: buffer, contentType: writer.FormDataContentType()}, nil
}

func writeFormDataRow(writer *multipart.Writer, row entity.FormDataRow) *entity.SendError {
	if !row.IsFile() {
		_ = writer.WriteField(row.Key, row.Value)
		return nil
	}
	content, sendErr := readFile(row.Src.Src)
	if sendErr != nil {
		return sendErr
	}
	part, _ := writer.CreateFormFile(row.Key, filepath.Base(row.Src.Src))
	_, _ = part.Write(content)
	return nil
}

func readFile(path string) ([]byte, *entity.SendError) {
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		sendErr := entity.NewSendError(senderrorkind.MissingFile, path)
		return nil, &sendErr
	}
	if err != nil {
		sendErr := entity.NewSendError(senderrorkind.Other, err.Error())
		return nil, &sendErr
	}
	return content, nil
}
