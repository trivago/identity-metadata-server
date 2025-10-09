package main

import (
	"net/http"

	"identity-metadata-server/internal/shared"
)

var (
	// ErrorNoClientCert is returned when a client certificate is required but
	// not provided.
	ErrorNoClientCert = shared.ErrorWithStatus{
		Message: "Client certificate required",
		Code:    http.StatusUnauthorized,
	}
	// ErrorNoIdentity is returned when a client certificate contains no identity.
	// The identity is expected to be in the email SAN.
	ErrorNoIdentity = shared.ErrorWithStatus{
		Message: "Certificate does not contain an identity",
		Code:    http.StatusBadRequest,
	}
	// ErrorNoEmail is returned when a client certificate does not contain an
	// origin restriction. The origin restrictions are expected to be in the
	// IP SANs.
	ErrorNoOrigins = shared.ErrorWithStatus{
		Message: "Certificate does not contain any origin constraints",
		Code:    http.StatusBadRequest,
	}
	// ErrorNoSerial is returned when a client certificate does not contain a
	// serial number.
	ErrorNoSerial = shared.ErrorWithStatus{
		Message: "Certificate does not contain a serial number",
		Code:    http.StatusBadRequest,
	}
	// ErrorCertificateExpired is returned when a client certificate has expired.
	ErrorCertificateExpired = shared.ErrorWithStatus{
		Message: "Certificate has expired",
		Code:    http.StatusForbidden,
	}
	// ErrorCertificateNotValidYet is returned when a client certificate is not
	// yet valid.
	ErrorCertificateNotValidYet = shared.ErrorWithStatus{
		Message: "Certificate not valid yet",
		Code:    http.StatusForbidden,
	}
	// ErrorUnknownTrustRoot is returned when a client certificate is not signed
	// by a known trust root.
	ErrorUnknownTrustRoot = shared.ErrorWithStatus{
		Message: "Certificate not signed by trust root",
		Code:    http.StatusForbidden,
	}
	// ErrorCertificateRevoked is returned when a client certificate has been
	// revoked.
	ErrorCertificateRevoked = shared.ErrorWithStatus{
		Message: "Certificate has been revoked",
		Code:    http.StatusGone,
	}
	// ErrorSigningKeyNotLoaded is returned when the server's signing key is not
	// loaded, but is required to sign a token.
	ErrorSigningKeyNotLoaded = shared.ErrorWithStatus{
		Message: "Signing key not loaded",
		Code:    http.StatusInternalServerError,
	}
	// ErrorNotAllowedForOrigin is returned when a token request is made from an
	// origin that is not allowed by the client certificate.
	ErrorNotAllowedForOrigin = shared.ErrorWithStatus{
		Message: "Token request not allowed for given origin",
		Code:    http.StatusForbidden,
	}
)
