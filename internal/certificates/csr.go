package certificates

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"identity-metadata-server/internal/shared"
	"net"
)

// CreateClientCSR generates a Certificate Signing Request (CSR) in PEM format
// using the provided PEM-encoded private key.
// It includes the given hostname as Common Name and DNS SAN, the email address,
// and the list of IP addresses in the appropriate SAN fields.
// The CSR is configured for mTLS client authentication usage.
func CreateClientCSR(privateKeyPEM []byte, hostname string, email string, ips []net.IP) ([]byte, error) {
	var (
		privateKey         any
		err                error
		signatureAlgorithm x509.SignatureAlgorithm
	)

	// Decode the PEM-encoded private key
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		signatureAlgorithm = x509.SHA256WithRSA
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		signatureAlgorithm = x509.ECDSAWithSHA256
		privateKey, err = x509.ParseECPrivateKey(block.Bytes)
	case "PRIVATE KEY":
		if privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
			switch privateKey.(type) {
			case *rsa.PrivateKey:
				signatureAlgorithm = x509.SHA256WithRSA
			case *ecdsa.PrivateKey:
				signatureAlgorithm = x509.ECDSAWithSHA256
			default:
				err = errors.New("unsupported PKCS#8 key type")
			}
		}
	default:
		err = errors.New("unsupported private key type")
	}
	if err != nil {
		return nil, errors.New("failed to parse private key: " + err.Error())
	}

	// Create the KeyUsage and ExtendedKeyUsage extensions
	usage, err := KeyUsageToExtension(x509.KeyUsageDigitalSignature)
	if err != nil {
		return nil, errors.New("failed to marshal key usage: " + err.Error())
	}

	extendedUsage, err := ExtKeyUsageToExtension([]x509.ExtKeyUsage{
		x509.ExtKeyUsageClientAuth,
	})
	if err != nil {
		return nil, errors.New("failed to marshal extended key usage: " + err.Error())
	}

	emailAddress := []string{}
	if email != "" {
		emailAddress = []string{email}
	}

	// Create the CSR template
	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   hostname,
			Organization: []string{"trivago"},
		},
		SignatureAlgorithm: signatureAlgorithm,
		DNSNames:           []string{hostname},
		EmailAddresses:     emailAddress,
		IPAddresses:        ips,

		ExtraExtensions: []pkix.Extension{
			usage,
			extendedUsage,
		},
	}

	// Create the CSR using the private key and template
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, errors.New("failed to create certificate request: " + err.Error())
	}

	// Encode CSR to PEM
	csrPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE REQUEST", Bytes: csrBytes})
	return csrPEM, nil
}

// CreateClientCSRFromCertificate generates a CSR from the given certificate
// using the provided PEM-encoded private key.
func CreateClientCSRFromCertificate(privateKeyPEM []byte, cert *x509.Certificate) ([]byte, error) {
	email := ""
	if len(cert.EmailAddresses) > 0 {
		email = cert.EmailAddresses[0]
	}
	return CreateClientCSR(privateKeyPEM, cert.Subject.CommonName, email, cert.IPAddresses)
}

// KeyUsageToExtension converts a x509.KeyUsage to a pkix.Extension.
// You can provide multiple KeyUsages by using a bitwise OR operation.
func KeyUsageToExtension(keyUsage x509.KeyUsage) (pkix.Extension, error) {
	return marshalKeyUsage(keyUsage)
}

// ExtKeyUsageToExtension converts a slice of x509.ExtKeyUsage to a pkix.Extension.
// If the ExtKeyUsage is empty, an empty pkix.Extension and no error is returned.
func ExtKeyUsageToExtension(extKeyUsage []x509.ExtKeyUsage) (pkix.Extension, error) {
	if len(extKeyUsage) == 0 {
		return pkix.Extension{}, nil
	}

	return marshalExtKeyUsage(extKeyUsage)
}

// HasKeyUsage checks if the CSR has exactly the keyusage specified as given.
func VerifyCSRKeyUsage(csr *x509.CertificateRequest, usage x509.KeyUsage) (bool, error) {
	ext := (*pkix.Extension)(nil)
	for _, e := range csr.Extensions {
		if e.Id.Equal(oidExtensionKeyUsage) {
			ext = &e
			break
		}
	}

	if ext == nil {
		return false, nil // Not found
	}

	if len(ext.Value) == 0 {
		return false, errors.New("malformed key usage value") // Malformed value
	}

	var cmpUsage [2]byte
	cmpUsage[0] = reverseBitsInAByte(byte(usage))
	cmpUsage[1] = reverseBitsInAByte(byte(usage >> 8))

	usageBitString := asn1.BitString{}
	if _, err := asn1.Unmarshal(ext.Value, &usageBitString); err != nil {
		return false, err // Marshalling error
	}

	if len(usageBitString.Bytes) == 0 {
		return false, errors.New("empty key usage bitstring")
	}

	if cmpUsage[0] != usageBitString.Bytes[0] {
		return false, nil // First byte mismatch
	}

	if usageBitString.BitLength > 8 && cmpUsage[1] != usageBitString.Bytes[1] {
		return false, nil // Second byte mismatch
	}

	return true, nil
}

// HasExtKeyUsage checks if the given CSR has all of and only the given usages.
func VerifyCSRExtKeyUsage(csr *x509.CertificateRequest, usage []x509.ExtKeyUsage) (bool, error) {
	usageAsOids := make([]asn1.ObjectIdentifier, len(usage))
	for i, u := range usage {
		if oid, ok := oidFromExtKeyUsage(u); ok {
			usageAsOids[i] = oid
		} else {
			return false, errors.New("unknown key usage to check for")
		}
	}

	cmpOIDs := func(a, b asn1.ObjectIdentifier) bool {
		return a.Equal(b)
	}

	foundAtLeastOne := false
	for _, ext := range csr.Extensions {
		if ext.Id.Equal(oidExtensionExtendedKeyUsage) {
			if len(ext.Value) == 0 {
				return false, errors.New("malformed extension value")
			}

			oids := []asn1.ObjectIdentifier{}
			if _, err := asn1.Unmarshal(ext.Value, &oids); err != nil {
				return false, err // Marhalling error
			}

			if len(oids) != len(usage) {
				return false, nil // Different number of usages
			}

			if !shared.EqualUnorderedFunc(oids, usageAsOids, cmpOIDs) {
				return false, nil // Different usages
			}
			foundAtLeastOne = true
		}
	}

	return foundAtLeastOne, nil // All checks passed
}
