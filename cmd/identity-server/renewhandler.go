package main

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"identity-metadata-server/internal/certificates"
	"identity-metadata-server/internal/shared"
	"io"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type RenewRequest struct {
	CSR string `json:"csr"`
}

func HandleRenewRequest(c *gin.Context, crl *CertificateRevocationList, cfg certificates.GCPCertificateAuthorityConfig) {
	// Check if we can get a client from the context
	client, err := NewClientFromContext(c, crl)
	if err != nil {
		log.Error().Err(err).Str("clientIP", c.ClientIP()).Msg("Failed to validate client identity")
		shared.HttpError(c, http.StatusInternalServerError, err)
		return
	}

	request := RenewRequest{}

	// Read the body either as JSON or as a PEM file
	if c.ContentType() == "application/x-pem-file" {
		csrData, err := io.ReadAll(c.Request.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read CSR from request")
			shared.HttpError(c, http.StatusBadRequest, err)
			return
		}
		request.CSR = string(csrData)
	} else if err := c.BindJSON(&request); err != nil {
		log.Error().Err(err).Msg("Failed to parse request")
		shared.HttpError(c, http.StatusBadRequest, err)
		return
	}

	pemBlock, _ := pem.Decode([]byte(request.CSR))
	if pemBlock == nil {
		err = fmt.Errorf("CSR was not a valid PEM block")
		log.Error().Err(err).Msg("Failed to decode request")
		shared.HttpError(c, http.StatusBadRequest, err)
		return
	}

	csr, err := x509.ParseCertificateRequest(pemBlock.Bytes)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse CSR")
		shared.HttpError(c, http.StatusBadRequest, err)
		return
	}

	if err := csr.CheckSignature(); err != nil {
		log.Error().Err(err).Msg("CSR signature verification failed")
		shared.HttpError(c, http.StatusBadRequest, errors.Join(err, fmt.Errorf("CSR signature invalid")))
		return
	}

	// Verify the CSR doing a refresh, not a new request
	if err := VerifyRenewRequest(csr, client.Certificate); err != nil {
		log.Error().Err(err).Msg("CSR is not a valid refresh request")
		shared.HttpError(c, http.StatusUnprocessableEntity, err)
		return
	}

	// Get a token to access the GCP Certificate Authority
	bearerToken, err := GetIdentityServerToken([]string{certificateAuthorityScope}, c.Request.Context())
	if err != nil {
		shared.HttpError(c, http.StatusInternalServerError, err)
		return
	}

	// Create a new certificate from the CSR
	lifetime := viper.GetDuration("server.certAuthority.clientCertLifetime")

	cert, err := certificates.CreateGCPCertificateFromCSR(cfg, bearerToken, []byte(request.CSR), lifetime, c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to create certificate from CSR")
		shared.HttpError(c, http.StatusInternalServerError, err)
		return
	}

	// Return the cert to the client
	certPEM, err := certificates.EncodeCertificateToPEM(cert)
	if err != nil {
		log.Error().Err(err).Msg("failed to encode certificate to PEM")
		shared.HttpError(c, http.StatusInternalServerError, err)
		return
	}

	c.Header("Content-Type", "application/x-pem-file")
	c.Header("Content-Disposition", "attachment; filename=client.cert")
	c.String(http.StatusOK, string(certPEM))
}

// VerifyRenewRequest checks if the CSR is a valid refresh request for the current certificate.
// If the request is valid, it returns nil, otherwise it returns an HTTP compatible error.
func VerifyRenewRequest(csr *x509.CertificateRequest, cert *x509.Certificate) error {

	// Check if the CSR is a valid refresh of the current certificate
	// Check 1: Hostname
	if len(csr.DNSNames) != 1 {
		return shared.NewErrorWithStatus(http.StatusUnprocessableEntity, "CSR must contain exactly one DNS name")
	}

	if csr.DNSNames[0] != cert.Subject.CommonName {
		return shared.NewErrorWithStatus(http.StatusUnprocessableEntity, "CSR DNS name does not match current client certificate")
	}

	if csr.Subject.CommonName != cert.Subject.CommonName {
		return shared.NewErrorWithStatus(http.StatusUnprocessableEntity, "CSR Common Name does not match current client certificate")
	}

	// Check 2: Email address (identity)
	if !shared.EqualUnordered(csr.EmailAddresses, cert.EmailAddresses) {
		return shared.NewErrorWithStatus(http.StatusUnprocessableEntity, "CSR email address does not match current client certificate")
	}

	// Check 3: IP addresses (origin)
	IpEqual := func(a, b net.IP) bool {
		return a.Equal(b)
	}
	if !shared.EqualUnorderedFunc(csr.IPAddresses, cert.IPAddresses, IpEqual) {
		return shared.NewErrorWithStatus(http.StatusUnprocessableEntity, "CSR IP addresses do not match current client certificate")
	}

	// Check 4: Key usage
	if ok, err := certificates.VerifyCSRKeyUsage(csr, x509.KeyUsageDigitalSignature); !ok {
		err = errors.Join(err, fmt.Errorf("key usage validation failed"))
		return shared.WrapErrorWithStatus(err, http.StatusUnprocessableEntity)
	}

	if ok, err := certificates.VerifyCSRExtKeyUsage(csr, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}); !ok {
		err = errors.Join(err, fmt.Errorf("extended key usage validation failed"))
		return shared.WrapErrorWithStatus(err, http.StatusUnprocessableEntity)
	}

	return nil
}
