package main

import (
	"crypto/x509"
	"encoding/hex"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// IdentityClient represents a client that has registered with the identity server.
type IdentityClient struct {
	// Host is the hostname of the client. This is used to identify the client
	// and must be unique. This string is used as the JWT subject.
	Host string
	// Identity is the identity to bind to. This is expected to be a GCP service
	// account email address.
	Identity string
	// IpAddr is the list of IP addresses the client is allowed to connect from.
	IpAddr []net.IP
	// Certificate is the certificate used to authenticate the client.
	Certificate *x509.Certificate
	// SerialNumber is the serial number of the certificate as hex string.
	SerialNumber string
}

// NewClientFromCert creates a new IdentityClient from a client certificate.
func NewClientFromCert(cert *x509.Certificate) (*IdentityClient, error) {
	if len(cert.EmailAddresses) == 0 {
		log.Error().Str("identity", cert.Subject.CommonName).Msg("Missing email address in certificate")
		return nil, ErrorNoIdentity
	}

	if len(cert.IPAddresses) == 0 {
		log.Error().Str("identity", cert.Subject.CommonName).Msg("Missing IP address(es) in certificate")
		return nil, ErrorNoOrigins
	}

	if cert.SerialNumber == nil {
		log.Error().Str("identity", cert.Subject.CommonName).Msg("Missing serial in certificate")
		return nil, ErrorNoSerial
	}

	return &IdentityClient{
		Host:         strings.ToLower(cert.Subject.CommonName),
		Identity:     cert.EmailAddresses[0],
		IpAddr:       cert.IPAddresses,
		Certificate:  cert,
		SerialNumber: hex.EncodeToString(cert.SerialNumber.Bytes()),
	}, nil
}

// NewClientFromContext creates a new IdentityClient from the given gin.Context.
// It extracts the client certificate from the context and verifies it.
// If the certificate is valid, and the client is allowed to connect from the
// client IP address, it returns a new IdentityClient.
func NewClientFromContext(c *gin.Context, crl *CertificateRevocationList) (*IdentityClient, error) {
	if len(c.Request.TLS.PeerCertificates) < 1 {
		return nil, ErrorNoClientCert
	}

	clientCert := c.Request.TLS.PeerCertificates[0]
	client, err := NewClientFromCert(clientCert)
	if err != nil {
		return nil, err
	}

	originIP := net.ParseIP(c.ClientIP())
	if !client.IsFromValidOrigin(originIP) {
		log.Error().
			Str("client", c.ClientIP()).
			Str("identity", client.Identity).
			Msg("access request from invalid origin")
		return nil, ErrorNotAllowedForOrigin
	}

	if err := client.VerifyCertificate(crl); err != nil {
		return nil, err
	}

	return client, nil
}

// VerifyCertificate verifies the client certificate.
// If no error is returned, the certificate is valid.
func (client *IdentityClient) VerifyCertificate(crl *CertificateRevocationList) error {
	if client.Certificate == nil {
		return ErrorNoClientCert
	}

	if len(client.SerialNumber) == 0 {
		return ErrorNoSerial
	}

	now := time.Now()
	if now.Before(client.Certificate.NotBefore) {
		return ErrorCertificateNotValidYet
	}

	if now.After(client.Certificate.NotAfter) {
		return ErrorCertificateExpired
	}

	if !crl.IsCertFromPool(client.Certificate) {
		return ErrorUnknownTrustRoot
	}

	if crl.IsSerialRevoked(client.SerialNumber) {
		return ErrorCertificateRevoked
	}

	return nil
}

// IsFromValidOrigin checks if the given IP address is in the list of allowed IP addresses.
func (client *IdentityClient) IsFromValidOrigin(ip net.IP) bool {
	for _, allowedIP := range client.IpAddr {
		if allowedIP.Equal(ip) {
			return true
		}
	}
	return false
}
