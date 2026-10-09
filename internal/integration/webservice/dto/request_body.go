package dto

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/bodymode"
)

type RequestBody struct {
	Content     io.Reader `json:"-"`
	ContentType string    `json:"content_type"`
}

func NewRequestBody(body entity.Body, files map[string][]byte) RequestBody {
	switch body.Mode {
	case bodymode.Raw:
		return RequestBody{Content: strings.NewReader(body.Raw.Content)}
	case bodymode.UrlEncoded:
		return RequestBody{Content: strings.NewReader(urlEncoded(body.UrlEncoded))}
	case bodymode.GraphQl:
		payload, _ := json.Marshal(NewGraphQlBody(body.GraphQl))
		return RequestBody{Content: bytes.NewReader(payload)}
	case bodymode.Binary:
		return RequestBody{Content: bytes.NewReader(files[body.Binary.Src])}
	case bodymode.FormData:
		return formDataBody(body.FormData, files)
	default:
		return RequestBody{}
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

func formDataBody(rows []entity.FormDataRow, files map[string][]byte) RequestBody {
	buffer := &bytes.Buffer{}
	writer := multipart.NewWriter(buffer)
	for _, row := range rows {
		if row.IsActive() {
			writeFormDataRow(writer, row, files)
		}
	}
	_ = writer.Close()
	return RequestBody{Content: buffer, ContentType: writer.FormDataContentType()}
}

func writeFormDataRow(writer *multipart.Writer, row entity.FormDataRow, files map[string][]byte) {
	if !row.IsFile() {
		_ = writer.WriteField(row.Key, row.Value)
		return
	}
	part, _ := writer.CreateFormFile(row.Key, filepath.Base(row.Src.Src))
	_, _ = part.Write(files[row.Src.Src])
}
