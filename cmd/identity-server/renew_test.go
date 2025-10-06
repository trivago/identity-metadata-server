package main

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"identity-metadata-server/internal/certificates"
	"math/big"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyRenewRequest(t *testing.T) {
	assert := assert.New(t)

	key, err := certificates.CreateECPrivateKeyPEM(certificates.KeyStrengthNormal)
	assert.NoError(err)
	assert.NotNil(key)

	clientIPs := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}

	dummyCert := CreateDummyCertificate("test", "test@test", clientIPs)

	csrPEM, err := certificates.CreateClientCSR(key, "test", "test@test", clientIPs)
	assert.NoError(err)
	assert.NotNil(csrPEM)

	csrData, _ := pem.Decode(csrPEM)
	assert.NotNil(csrData)

	csr, err := x509.ParseCertificateRequest(csrData.Bytes)
	assert.NoError(err)
	assert.NotNil(csr)

	err = VerifyRenewRequest(csr, dummyCert)
	assert.NoError(err)
}

func TestVerifyRenewRequestSAChange(t *testing.T) {
	assert := assert.New(t)

	key, err := certificates.CreateECPrivateKeyPEM(certificates.KeyStrengthNormal)
	assert.NoError(err)
	assert.NotNil(key)

	clientIPs := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}

	dummyCert := CreateDummyCertificate("test", "test@test", clientIPs)

	csrPEM, err := certificates.CreateClientCSR(key, "test", "pivot@test", clientIPs)
	assert.NoError(err)
	assert.NotNil(csrPEM)

	csrData, _ := pem.Decode(csrPEM)
	assert.NotNil(csrData)

	csr, err := x509.ParseCertificateRequest(csrData.Bytes)
	assert.NoError(err)
	assert.NotNil(csr)

	err = VerifyRenewRequest(csr, dummyCert)
	assert.Error(err)
}

func TestVerifyRenewRequestIPChange(t *testing.T) {
	assert := assert.New(t)

	key, err := certificates.CreateECPrivateKeyPEM(certificates.KeyStrengthNormal)
	assert.NoError(err)
	assert.NotNil(key)

	clientIPs := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	newClientIPs := []net.IP{net.ParseIP("192.168.178.1"), net.ParseIP("::1")}

	dummyCert := CreateDummyCertificate("test", "test@test", clientIPs)

	csrPEM, err := certificates.CreateClientCSR(key, "test", "test@test", newClientIPs)
	assert.NoError(err)
	assert.NotNil(csrPEM)

	csrData, _ := pem.Decode(csrPEM)
	assert.NotNil(csrData)

	csr, err := x509.ParseCertificateRequest(csrData.Bytes)
	assert.NoError(err)
	assert.NotNil(csr)

	err = VerifyRenewRequest(csr, dummyCert)
	assert.Error(err)
}

func TestVerifyRenewRequestIdentityChange(t *testing.T) {
	assert := assert.New(t)

	key, err := certificates.CreateECPrivateKeyPEM(certificates.KeyStrengthNormal)
	assert.NoError(err)
	assert.NotNil(key)

	clientIPs := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}

	dummyCert := CreateDummyCertificate("test", "test@test", clientIPs)

	csrPEM, err := certificates.CreateClientCSR(key, "hacker", "test@test", clientIPs)
	assert.NoError(err)
	assert.NotNil(csrPEM)

	csrData, _ := pem.Decode(csrPEM)
	assert.NotNil(csrData)

	csr, err := x509.ParseCertificateRequest(csrData.Bytes)
	assert.NoError(err)
	assert.NotNil(csr)

	err = VerifyRenewRequest(csr, dummyCert)
	assert.Error(err)
}

func CreateDummyCertificate(hostname string, email string, ips []net.IP) *x509.Certificate {
	return &x509.Certificate{
		Subject: pkix.Name{
			CommonName: hostname,
		},
		DNSNames:       []string{hostname},
		EmailAddresses: []string{email},
		IPAddresses:    ips,
		SerialNumber:   big.NewInt(1),
		KeyUsage:       x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}
}
