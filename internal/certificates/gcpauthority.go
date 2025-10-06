package certificates

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"hash/fnv"
	"identity-metadata-server/internal/shared"
	"regexp"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

// GCPCertificateAuthorityConfig is used to reference a GCP Certificate Authority.
type GCPCertificateAuthorityConfig struct {
	ProjectID            string
	Location             string
	CertificatePool      string
	CertificateAuthority string
}

// See https://cloud.google.com/certificate-authority-service/docs/reference/rest/v1/projects.locations.caPools.certificates#Certificate
type GCPCertificate struct {
	Name             string `json:"name"`
	Lifetime         string `json:"lifetime"`
	CertificateAsPEM string `json:"pemCertificate,omitempty"`
	CSR              string `json:"pemCsr,omitempty"`
}

// GetGCPCertificate retrieves a certificate from GCP Certificate Authority
// using the provided access token and certificate id.
// It returns the certificate as an x509.Certificate object.
func GetGCPCertificate(config GCPCertificateAuthorityConfig, gcpAccessToken string, certificateId string, ctx context.Context) (*x509.Certificate, error) {
	// https://cloud.google.com/certificate-authority-service/docs/reference/rest/v1/projects.locations.caPools.certificates/get
	requestURL := fmt.Sprintf(
		"https://privateca.googleapis.com/v1/projects/%s/locations/%s/caPools/%s/certificates/%s",
		config.ProjectID,
		config.Location,
		config.CertificatePool,
		certificateId,
	)

	log.Debug().Str("requestURL", requestURL).Msg("Requesting GCP certificate")

	response, err := shared.HttpGETJson[GCPCertificate](
		requestURL,
		nil,
		map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + gcpAccessToken,
		},
		nil, ctx)

	if err != nil {
		return nil, err
	}

	if response.CertificateAsPEM == "" {
		return nil, fmt.Errorf("empty certificate returned")
	}

	rawCertBlock, _ := pem.Decode([]byte(response.CertificateAsPEM))
	if rawCertBlock == nil {
		return nil, fmt.Errorf("returned certificate was not a valid PEM block")
	}

	return x509.ParseCertificate(rawCertBlock.Bytes)
}

// CreateGCPCertificateFromCSR creates a certificate from a CSR using the GCP Certificate Authority.
// The certificate id is generated from the CSR's Common Name and a hash of the CSR.
// If the certificate for the given CSR already exists, it is returned.
// The provided gcpAccessToken is used to authenticate the request.
// The lifetime parameter specifies the desired lifetime of the certificate.
func CreateGCPCertificateFromCSR(config GCPCertificateAuthorityConfig, gcpAccessToken string, csrPEM []byte, lifetime time.Duration, ctx context.Context) (*x509.Certificate, error) {
	if lifetime <= 0 {
		return nil, fmt.Errorf("lifetime must be greater than 0")
	}

	rawCSRBlock, _ := pem.Decode([]byte(csrPEM))
	if rawCSRBlock == nil {
		return nil, fmt.Errorf("CSR was not a valid PEM block")
	}

	// Get the CN from the CSR
	csr, err := x509.ParseCertificateRequest(rawCSRBlock.Bytes)
	if err != nil {
		return nil, err
	}

	// We need a suffix that is stable for the same CSR.
	// No need to use a cryptographic hash as the CSR is useless without the private key.
	certHash := fnv.New32a()
	if _, err = certHash.Write(csr.Raw); err != nil {
		return nil, err
	}

	// Certificate names must be less than 63 characters.
	// We use a 32bit = 4 byte hash, i.e. 8 hex characters and a dash, so the suffix is 9 characters.
	hostname := strings.ToLower(csr.Subject.CommonName)
	maxHostNameLength := 63 - 9
	if len(hostname) >= maxHostNameLength {
		hostname = hostname[:maxHostNameLength]
	}
	// Make sure the hostname is valid for GCP.
	sanitizedHost := regexp.MustCompile(`[^a-z0-9_-]`).ReplaceAllString(hostname, "-")
	if len(sanitizedHost) > maxHostNameLength {
		sanitizedHost = sanitizedHost[:maxHostNameLength]
	}
	certificateId := fmt.Sprintf("%s-%x", sanitizedHost, certHash.Sum32())

	// If the certificate already exists, we return it.
	existingCert, err := GetGCPCertificate(config, gcpAccessToken, certificateId, ctx)
	if existingCert != nil {
		return existingCert, err
	}

	logEx := log.Debug().Err(err)
	if httpErr, ok := err.(shared.ErrorWithStatus); ok {
		logEx = logEx.Int("status", httpErr.Code)
	}
	logEx.Msg("Certificate not found, creating new certificate")

	request := GCPCertificate{
		Lifetime: strconv.FormatInt(int64(lifetime.Seconds()), 10) + "s",
		CSR:      string(csrPEM),
	}

	requestBody, err := jsoniter.Marshal(request)
	if err != nil {
		return nil, err
	}

	// https://cloud.google.com/certificate-authority-service/docs/reference/rest/v1/projects.locations.caPools.certificates/create
	requestURL := fmt.Sprintf(
		"https://privateca.googleapis.com/v1/projects/%s/locations/%s/caPools/%s/certificates?certificateId=%s",
		config.ProjectID,
		config.Location,
		config.CertificatePool,
		certificateId,
	)

	log.Debug().Str("requestURL", requestURL).Msg("Requesting GCP certificate")

	// Call the GCP certificate authority service to create a certificate
	// using the provided CSR and access token.
	response, err := shared.HttpPOSTJson[GCPCertificate](
		requestURL, requestBody,
		map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + gcpAccessToken,
		},
		nil, ctx)

	if err != nil {
		return nil, err
	}
	if len(response.CertificateAsPEM) == 0 {
		return nil, fmt.Errorf("empty certificate returned")
	}

	rawCertBlock, _ := pem.Decode([]byte(response.CertificateAsPEM))
	if rawCertBlock == nil {
		return nil, fmt.Errorf("returned certificate was not a valid PEM block")
	}

	return x509.ParseCertificate(rawCertBlock.Bytes)
}
