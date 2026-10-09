package model

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

type HistoryResult struct {
	StatusText   string `json:"status_text"`
	ErrorKind    string `json:"error_kind"`
	ErrorMessage string `json:"error_message"`
	StatusCode   int    `json:"status_code"`
	TotalMs      int64  `json:"total_ms"`
	SizeBytes    int64  `json:"size_bytes"`
}

func NewHistoryResult(result entity.HistoryResult) HistoryResult {
	return HistoryResult{
		StatusCode:   result.StatusCode,
		StatusText:   result.StatusText,
		ErrorKind:    result.ErrorKind.String(),
		ErrorMessage: result.ErrorMessage,
		TotalMs:      result.Total.Milliseconds(),
		SizeBytes:    result.SizeBytes,
	}
}

func (h *HistoryResult) ToEntity() entity.HistoryResult {
	result := entity.HistoryResult{
		StatusCode:   h.StatusCode,
		StatusText:   h.StatusText,
		ErrorMessage: h.ErrorMessage,
		Total:        time.Duration(h.TotalMs) * time.Millisecond,
		SizeBytes:    h.SizeBytes,
	}
	if h.ErrorKind != "" {
		result.ErrorKind = senderrorkind.GetSendErrorKind(h.ErrorKind)
	}
	return result
}
