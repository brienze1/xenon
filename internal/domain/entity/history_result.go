package entity

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/enums"
)

type HistoryResult struct {
	ErrorKind    enums.SendErrorKind
	StatusText   string
	ErrorMessage string
	Total        time.Duration
	StatusCode   int
	SizeBytes    int64
}

func NewHistoryResult(response Response) HistoryResult {
	result := HistoryResult{
		StatusCode: response.StatusCode,
		StatusText: response.StatusText,
		Total:      response.Timings.Total,
		SizeBytes:  response.SizeBytes,
	}
	if !response.IsOk() {
		result.ErrorKind = response.Error.Kind
		result.ErrorMessage = response.Error.Message
	}
	return result
}

func (h HistoryResult) IsOk() bool {
	return h.ErrorKind == ""
}
