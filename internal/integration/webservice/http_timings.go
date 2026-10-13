package webservice

import (
	"crypto/tls"
	"net/http/httptrace"
	"sync"
	"time"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type timingTrace struct {
	start        time.Time
	dnsStart     time.Time
	connectStart time.Time
	tlsStart     time.Time
	wroteRequest time.Time
	firstByte    time.Time
	timings      entity.Timings
	mu           sync.Mutex
}

func newTimingTrace() *timingTrace {
	return &timingTrace{start: time.Now()}
}

func (t *timingTrace) clientTrace() *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart:             func(httptrace.DNSStartInfo) { t.mark(&t.dnsStart) },
		DNSDone:              func(httptrace.DNSDoneInfo) { t.measure(&t.timings.Dns, t.dnsStart) },
		ConnectStart:         func(string, string) { t.mark(&t.connectStart) },
		ConnectDone:          func(string, string, error) { t.measure(&t.timings.Connect, t.connectStart) },
		TLSHandshakeStart:    func() { t.mark(&t.tlsStart) },
		TLSHandshakeDone:     func(tls.ConnectionState, error) { t.measure(&t.timings.Tls, t.tlsStart) },
		WroteRequest:         func(httptrace.WroteRequestInfo) { t.mark(&t.wroteRequest) },
		GotFirstResponseByte: func() { t.firstResponseByte() },
	}
}

func (t *timingTrace) mark(moment *time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	*moment = time.Now()
}

func (t *timingTrace) measure(duration *time.Duration, since time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	*duration = time.Since(since)
}

func (t *timingTrace) firstResponseByte() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.firstByte = time.Now()
	t.timings.FirstByte = t.firstByte.Sub(t.wroteRequest)
}

func (t *timingTrace) finish() entity.Timings {
	t.mu.Lock()
	defer t.mu.Unlock()
	end := time.Now()
	if !t.firstByte.IsZero() {
		t.timings.Download = end.Sub(t.firstByte)
	}
	t.timings.Total = end.Sub(t.start)
	return t.timings
}
