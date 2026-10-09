package entity

import (
	"testing"

	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

func TestNewSendError(t *testing.T) {
	cases := []struct {
		got  SendError
		want SendError
	}{
		{NewSendError(senderrorkind.Dns, "x.test"), SendError{Kind: senderrorkind.Dns, Message: "Could not resolve host x.test"}},
		{NewSendError(senderrorkind.ConnectionRefused, "127.0.0.1:1"), SendError{Kind: senderrorkind.ConnectionRefused, Message: "Could not connect to 127.0.0.1:1 (connection refused)"}},
		{NewSendError(senderrorkind.Timeout, "100"), SendError{Kind: senderrorkind.Timeout, Message: "Timed out after 100 ms", Hint: "Increase the timeout in the request Settings (0 = no timeout)"}},
		{NewSendError(senderrorkind.Tls, ""), SendError{Kind: senderrorkind.Tls, Message: "Certificate not trusted", Hint: "Turn off SSL certificate verification in Settings to send anyway"}},
		{NewSendError(senderrorkind.Tls, "bad record"), SendError{Kind: senderrorkind.Tls, Message: "TLS handshake failed: bad record", Hint: "Turn off SSL certificate verification in Settings to send anyway"}},
		{NewSendError(senderrorkind.Cancelled, "ignored"), SendError{Kind: senderrorkind.Cancelled, Message: "Request cancelled"}},
		{NewSendError(senderrorkind.MissingFile, "/a.bin"), SendError{Kind: senderrorkind.MissingFile, Message: "File not found: /a.bin", Hint: "Choose the file again in the Body tab"}},
		{NewSendError(senderrorkind.InvalidUrl, ""), SendError{Kind: senderrorkind.InvalidUrl, Message: "Enter a URL to send the request", Hint: "Check the URL"}},
		{NewSendError(senderrorkind.InvalidUrl, "ftp://x"), SendError{Kind: senderrorkind.InvalidUrl, Message: "Invalid URL: ftp://x", Hint: "Check the URL"}},
		{NewSendError(senderrorkind.Other, "boom"), SendError{Kind: senderrorkind.Other, Message: "boom"}},
		{NewSendError("weird", "boom"), SendError{Kind: senderrorkind.Other, Message: "boom"}},
	}
	for _, tc := range cases {
		if tc.got != tc.want {
			t.Errorf("got %+v, want %+v", tc.got, tc.want)
		}
	}
}
