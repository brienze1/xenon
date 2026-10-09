package entity

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

const (
	timeoutHint     = "Increase the timeout in the request Settings (0 = no timeout)"
	tlsHint         = "Turn off SSL certificate verification in Settings to send anyway"
	missingFileHint = "Choose the file again in the Body tab"
	invalidUrlHint  = "Check the URL"
)

type SendError struct {
	Kind    enums.SendErrorKind
	Message string
	Hint    string
}

func NewSendError(kind enums.SendErrorKind, detail string) SendError {
	switch kind {
	case senderrorkind.Dns:
		return SendError{Kind: kind, Message: "Could not resolve host " + detail}
	case senderrorkind.ConnectionRefused:
		return SendError{Kind: kind, Message: "Could not connect to " + detail + " (connection refused)"}
	case senderrorkind.Timeout:
		return SendError{Kind: kind, Message: "Timed out after " + detail + " ms", Hint: timeoutHint}
	case senderrorkind.Tls:
		return SendError{Kind: kind, Message: tlsMessage(detail), Hint: tlsHint}
	case senderrorkind.Cancelled:
		return SendError{Kind: kind, Message: "Request cancelled"}
	case senderrorkind.MissingFile:
		return SendError{Kind: kind, Message: "File not found: " + detail, Hint: missingFileHint}
	case senderrorkind.InvalidUrl:
		return SendError{Kind: kind, Message: invalidUrlMessage(detail), Hint: invalidUrlHint}
	default:
		return SendError{Kind: senderrorkind.Other, Message: detail}
	}
}

func tlsMessage(detail string) string {
	if detail == "" {
		return "Certificate not trusted"
	}
	return "TLS handshake failed: " + detail
}

func invalidUrlMessage(detail string) string {
	if detail == "" {
		return "Enter a URL to send the request"
	}
	return "Invalid URL: " + detail
}
