package entity

import (
	"strconv"
	"time"
)

type RequestSettings struct {
	TimeoutMs        int
	FollowRedirects  bool
	VerifySsl        bool
	EncodeUrl        bool
	DisableCookieJar bool
}

func DefaultRequestSettings() RequestSettings {
	return RequestSettings{FollowRedirects: true, VerifySsl: true, EncodeUrl: true}
}

func (r RequestSettings) HasTimeout() bool {
	return r.TimeoutMs > 0
}

func (r RequestSettings) TimeoutText() string {
	return strconv.Itoa(r.TimeoutMs)
}

func (r RequestSettings) Timeout() time.Duration {
	return time.Duration(r.TimeoutMs) * time.Millisecond
}
