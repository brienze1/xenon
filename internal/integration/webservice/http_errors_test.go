package webservice

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"testing"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

type timeoutError struct{ timeout bool }

func (t timeoutError) Error() string   { return "net" }
func (t timeoutError) Timeout() bool   { return t.timeout }
func (t timeoutError) Temporary() bool { return false }

func TestSendErrorKinds(t *testing.T) {
	target, _ := url.Parse("http://127.0.0.1:9")
	request := entity.EffectiveRequest{Settings: entity.RequestSettings{TimeoutMs: 75}}
	wrap := func(err error) error { return &url.Error{Op: "Get", URL: target.String(), Err: err} }
	cases := []struct {
		err     error
		kind    enums.SendErrorKind
		message string
	}{
		{wrap(context.Canceled), senderrorkind.Cancelled, "Request cancelled"},
		{wrap(context.DeadlineExceeded), senderrorkind.Timeout, "Timed out after 75 ms"},
		{wrap(timeoutError{timeout: true}), senderrorkind.Timeout, "Timed out after 75 ms"},
		{wrap(&net.DNSError{Name: "x.invalid", IsTimeout: true}), senderrorkind.Timeout, "Timed out after 75 ms"},
		{wrap(&net.DNSError{Name: "x.invalid", IsNotFound: true}), senderrorkind.Dns, "Could not resolve host x.invalid"},
		{wrap(&net.OpError{Op: "dial", Err: os.NewSyscallError("connect", connectionRefused)}), senderrorkind.ConnectionRefused, "Could not connect to 127.0.0.1:9 (connection refused)"},
		{wrap(&tls.CertificateVerificationError{Err: errors.New("bad")}), senderrorkind.Tls, "Certificate not trusted"},
		{wrap(x509.UnknownAuthorityError{}), senderrorkind.Tls, "Certificate not trusted"},
		{wrap(x509.HostnameError{Host: "x"}), senderrorkind.Tls, "Certificate not trusted"},
		{wrap(x509.CertificateInvalidError{}), senderrorkind.Tls, "Certificate not trusted"},
		{wrap(tls.RecordHeaderError{Msg: "first record does not look like a TLS handshake"}), senderrorkind.Tls, "TLS handshake failed: tls: first record does not look like a TLS handshake"},
		{wrap(tls.AlertError(40)), senderrorkind.Tls, "TLS handshake failed: tls: handshake failure"},
		{wrap(timeoutError{timeout: false}), senderrorkind.Other, "Get \"http://127.0.0.1:9\": net"},
		{fmt.Errorf("boom"), senderrorkind.Other, "boom"},
	}
	for _, tc := range cases {
		got := sendError(tc.err, request, target)
		if got.Kind != tc.kind || got.Message != tc.message {
			t.Errorf("%v: got %s %q, want %s %q", tc.err, got.Kind, got.Message, tc.kind, tc.message)
		}
	}
}
