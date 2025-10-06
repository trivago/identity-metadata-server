package main

import (
	"net/http"

	"identity-metadata-server/internal/shared"
)

var (
	ErrorNoClientCert = shared.ErrorWithStatus{
		Message: "Client certificate required",
		Code:    http.StatusUnauthorized,
	}
	ErrorNoIdentity = shared.ErrorWithStatus{
		Message: "Certificate does not contain an identity",
		Code:    http.StatusBadRequest,
	}
	ErrorNoOrigins = shared.ErrorWithStatus{
		Message: "Certificate does not contain any origin constraints",
		Code:    http.StatusBadRequest,
	}
	ErrorNoSerial = shared.ErrorWithStatus{
		Message: "Certificate does not contain a serial number",
		Code:    http.StatusBadRequest,
	}
	ErrorCertificateExpired = shared.ErrorWithStatus{
		Message: "Certificate has expired",
		Code:    http.StatusForbidden,
	}
	ErrorCertificateNotValidYet = shared.ErrorWithStatus{
		Message: "Certificate not valid yet",
		Code:    http.StatusForbidden,
	}
	ErrorUnknownTrustRoot = shared.ErrorWithStatus{
		Message: "Certificate not signed by trust root",
		Code:    http.StatusForbidden,
	}
	ErrorCertificateRevoked = shared.ErrorWithStatus{
		Message: "Certificate has been revoked",
		Code:    http.StatusGone,
	}
	ErrorSigningKeyNotLoaded = shared.ErrorWithStatus{
		Message: "Signing key not loaded",
		Code:    http.StatusInternalServerError,
	}
	ErrorNotAllowedForOrigin = shared.ErrorWithStatus{
		Message: "Token request not allowed for given origin",
		Code:    http.StatusForbidden,
	}
)
