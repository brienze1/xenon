package webservice

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/url"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/senderrorkind"
)

func sendError(err error, request entity.EffectiveRequest, target *url.URL) entity.SendError {
	var dnsError *net.DNSError
	var certificateError *tls.CertificateVerificationError
	var unknownAuthority x509.UnknownAuthorityError
	var hostnameError x509.HostnameError
	var invalidCertificate x509.CertificateInvalidError
	var recordHeaderError tls.RecordHeaderError
	var alertError tls.AlertError
	switch {
	case errors.Is(err, context.Canceled):
		return entity.NewSendError(senderrorkind.Cancelled, "")
	case isTimeout(err):
		return entity.NewSendError(senderrorkind.Timeout, request.TimeoutText())
	case errors.As(err, &dnsError):
		return entity.NewSendError(senderrorkind.Dns, dnsError.Name)
	case errors.Is(err, connectionRefused):
		return entity.NewSendError(senderrorkind.ConnectionRefused, target.Host)
	case errors.As(err, &certificateError), errors.As(err, &unknownAuthority), errors.As(err, &hostnameError), errors.As(err, &invalidCertificate):
		return entity.NewSendError(senderrorkind.Tls, "")
	case errors.As(err, &recordHeaderError):
		return entity.NewSendError(senderrorkind.Tls, recordHeaderError.Error())
	case errors.As(err, &alertError):
		return entity.NewSendError(senderrorkind.Tls, alertError.Error())
	default:
		return entity.NewSendError(senderrorkind.Other, err.Error())
	}
}

func isTimeout(err error) bool {
	var netError net.Error
	return errors.Is(err, context.DeadlineExceeded) || errors.As(err, &netError) && netError.Timeout()
}
