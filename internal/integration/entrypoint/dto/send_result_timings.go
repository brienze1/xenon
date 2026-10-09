package dto

import (
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type SendResultTimings struct {
	DnsMs       float64 `json:"dns_ms"`
	ConnectMs   float64 `json:"connect_ms"`
	TlsMs       float64 `json:"tls_ms"`
	FirstByteMs float64 `json:"first_byte_ms"`
	DownloadMs  float64 `json:"download_ms"`
	TotalMs     float64 `json:"total_ms"`
}

func NewSendResultTimings(timings entity.Timings) SendResultTimings {
	return SendResultTimings{
		DnsMs:       milliseconds(timings.Dns),
		ConnectMs:   milliseconds(timings.Connect),
		TlsMs:       milliseconds(timings.Tls),
		FirstByteMs: milliseconds(timings.FirstByte),
		DownloadMs:  milliseconds(timings.Download),
		TotalMs:     milliseconds(timings.Total),
	}
}

func milliseconds(duration time.Duration) float64 {
	return float64(duration.Microseconds()) / 1000
}
