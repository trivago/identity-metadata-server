package certificates

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"identity-metadata-server/internal/shared"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateClientCSR(t *testing.T) {
	assert := assert.New(t)

	key, err := CreateECPrivateKeyPEM(KeyStrengthNormal)
	assert.NoError(err)
	assert.NotNil(key)

	clientIPs := []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}

	csr, err := CreateClientCSR(key, "test", "test@test", clientIPs)
	assert.NoError(err)
	assert.NotNil(csr)

	block, _ := pem.Decode(csr)
	assert.Equal("CERTIFICATE REQUEST", block.Type)

	csrParsed, err := x509.ParseCertificateRequest(block.Bytes)
	assert.NoError(err)
	assert.NotNil(csrParsed)

	// Check if we set the expected fields

	keyUsageOk, err := VerifyCSRKeyUsage(csrParsed, x509.KeyUsageDigitalSignature)
	assert.NoError(err)
	assert.True(keyUsageOk)

	extKeyUsageok, err := VerifyCSRExtKeyUsage(csrParsed, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	assert.NoError(err)
	assert.True(extKeyUsageok)

	assert.Len(csrParsed.DNSNames, 1)
	assert.Len(csrParsed.EmailAddresses, 1)
	assert.Len(csrParsed.IPAddresses, 2)

	assert.Equal("test", csrParsed.Subject.CommonName)
	assert.Equal("test", csrParsed.DNSNames[0])
	assert.Equal("test@test", csrParsed.EmailAddresses[0])

	assert.True(shared.EqualUnorderedFunc(clientIPs, csrParsed.IPAddresses, func(a, b net.IP) bool {
		return a.Equal(b)
	}))
}

func TestVerifyCSRKeyUsage(t *testing.T) {
	assert := assert.New(t)

	// Exactly as expected

	usage, err := KeyUsageToExtension(x509.KeyUsageDigitalSignature)
	assert.NoError(err)

	ok, err := VerifyCSRKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, x509.KeyUsageDigitalSignature)

	assert.NoError(err)
	assert.True(ok)

	// Exactly as expected (2 values)

	usage, err = KeyUsageToExtension(x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment)
	assert.NoError(err)

	ok, err = VerifyCSRKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment)

	assert.NoError(err)
	assert.True(ok)

	// Exactly as expected (2 bytes)

	usage, err = KeyUsageToExtension(x509.KeyUsageDecipherOnly)
	assert.NoError(err)

	ok, err = VerifyCSRKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, x509.KeyUsageDecipherOnly)

	assert.NoError(err)
	assert.True(ok)

	// Exactly as expected (2 bytes, 2 values)

	usage, err = KeyUsageToExtension(x509.KeyUsageDigitalSignature | x509.KeyUsageDecipherOnly)
	assert.NoError(err)

	ok, err = VerifyCSRKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, x509.KeyUsageDigitalSignature|x509.KeyUsageDecipherOnly)

	assert.NoError(err)
	assert.True(ok)

	// More usages then expected

	usage, err = KeyUsageToExtension(x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment)
	assert.NoError(err)

	ok, err = VerifyCSRKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, x509.KeyUsageDigitalSignature)

	assert.NoError(err)
	assert.False(ok)

	// Less usages then expected

	usage, err = KeyUsageToExtension(x509.KeyUsageDigitalSignature)
	assert.NoError(err)

	ok, err = VerifyCSRKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, x509.KeyUsageDigitalSignature|x509.KeyUsageKeyEncipherment)

	assert.NoError(err)
	assert.False(ok)
}

func TestVerifyCSRExtKeyUsage(t *testing.T) {
	assert := assert.New(t)

	// Exactly as expected
	usage, err := ExtKeyUsageToExtension([]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	assert.NoError(err)
	ok, err := VerifyCSRExtKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	assert.NoError(err)
	assert.True(ok)

	// Exactly as expected (2 values)
	usage, err = ExtKeyUsageToExtension([]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny})
	assert.NoError(err)
	ok, err = VerifyCSRExtKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny})
	assert.NoError(err)
	assert.True(ok)

	// More than expected
	usage, err = ExtKeyUsageToExtension([]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	assert.NoError(err)
	ok, err = VerifyCSRExtKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny})
	assert.NoError(err)
	assert.False(ok)

	// Less than as expected (2 values)
	usage, err = ExtKeyUsageToExtension([]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageAny})
	assert.NoError(err)
	ok, err = VerifyCSRExtKeyUsage(&x509.CertificateRequest{
		Extensions: []pkix.Extension{usage},
	}, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
	assert.NoError(err)
	assert.False(ok)

}
