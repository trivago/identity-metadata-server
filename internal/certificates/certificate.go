package certificates

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// EncodeCertificateToPEM encodes an x509 certificate to PEM format.
func EncodeCertificateToPEM(cert *x509.Certificate) ([]byte, error) {
	if cert == nil {
		return nil, errors.New("certificate is nil")
	}

	pemData := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})

	if pemData == nil {
		return nil, errors.New("failed to encode certificate to PEM")
	}

	return pemData, nil
}
