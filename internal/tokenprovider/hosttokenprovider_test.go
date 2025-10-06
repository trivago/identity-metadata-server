package tokenprovider

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const (
	fileIdCACert     = "ca.cert"
	fileIdCAKey      = "ca.key"
	fileIdClientCert = "client.cert"
	fileIdClientKey  = "client.key"

	firstCertSerial = 1
	newCertSerial   = 42
)

type hostProviderTestContext struct {
	path map[string]string

	ca    *x509.Certificate
	caKey *rsa.PrivateKey
}

func (t *hostProviderTestContext) Add(name string, data []byte) error {
	tmpFile, err := os.CreateTemp("", name)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	t.path[name] = tmpFile.Name()

	if _, err := tmpFile.Write(data); err != nil {
		return err
	}
	return tmpFile.Sync()
}

func (t *hostProviderTestContext) Clean() {
	for _, path := range t.path {
		_ = os.Remove(path) // best-effort cleanup – ignore failure
	}
}

func NewMockIdentityServer(testContext *hostProviderTestContext) (*httptest.Server, error) {
	// Create a self-signed root certifcate
	var err error
	testContext.caKey, err = rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	caKeyPEM := new(bytes.Buffer)
	pem.Encode(caKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(testContext.caKey),
	})

	if err := testContext.Add(fileIdCAKey, caKeyPEM.Bytes()); err != nil {
		return nil, err
	}

	testContext.ca = &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization:  []string{"trivago"},
			Country:       []string{"DE"},
			Province:      []string{""},
			Locality:      []string{"Duesseldorf"},
			StreetAddress: []string{"Kesselstraße 5-7"},
			PostalCode:    []string{"40221"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour * 24),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, testContext.ca, testContext.ca, &testContext.caKey.PublicKey, testContext.caKey)
	if err != nil {
		return nil, err
	}

	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caDER,
	})

	if err := testContext.Add(fileIdCACert, caPEM.Bytes()); err != nil {
		return nil, err
	}

	// Create a server certificate
	serverKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, err
	}

	serverCert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"trivago"},
			Country:       []string{"DE"},
			Province:      []string{""},
			Locality:      []string{"Duesseldorf"},
			StreetAddress: []string{"Kesselstraße 5-7"},
			PostalCode:    []string{"40221"},
			CommonName:    "localhost",
		},
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour),
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature,
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, serverCert, testContext.ca, &serverKey.PublicKey, testContext.caKey)
	if err != nil {
		return nil, err
	}

	serverCertPEM := new(bytes.Buffer)
	pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertDER,
	})

	serverKeyPEM := new(bytes.Buffer)
	pem.Encode(serverKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverKey),
	})

	serverTLSCert, err := tls.X509KeyPair(serverCertPEM.Bytes(), serverKeyPEM.Bytes())
	if err != nil {
		return nil, err
	}

	router := gin.New()
	router.POST("/renew", func(c *gin.Context) {
		csrPEM, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusBadRequest, "failed to read CSR")
			return
		}

		certPEM, err := NewClientCertFromCSR(csrPEM, testContext.ca, testContext.caKey)
		if err != nil {
			c.String(http.StatusBadRequest, "failed to create client cert")
			return
		}
		c.Header("Content-Type", "application/x-pem-file")
		c.Header("Content-Disposition", "attachment; filename=cert.pem")

		c.String(http.StatusOK, string(certPEM))
	})

	// Return the client certificate serial number
	// This is used to verify that the new cert is being used
	router.GET("/identity", func(c *gin.Context) {
		if c.Request.TLS == nil || len(c.Request.TLS.PeerCertificates) == 0 {
			c.String(http.StatusBadRequest, "no client cert")
			return
		}
		c.String(http.StatusOK, c.Request.TLS.PeerCertificates[0].SerialNumber.String()+"\n")
	})

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caPEM.Bytes())

	testSrv := httptest.NewUnstartedServer(router)
	testSrv.EnableHTTP2 = true
	testSrv.TLS = &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{serverTLSCert},
		ClientAuth:   tls.VerifyClientCertIfGiven,
		ClientCAs:    certPool,
	}

	testSrv.StartTLS()
	return testSrv, nil
}

func NewClientCertFromCSR(csrPEM []byte, ca *x509.Certificate, caKey *rsa.PrivateKey) ([]byte, error) {
	csrDER, _ := pem.Decode(csrPEM)
	if csrDER == nil {
		return nil, errors.New("failed to parse CSR PEM")
	}

	csr, err := x509.ParseCertificateRequest(csrDER.Bytes)
	if err != nil {
		return nil, err
	}

	clientCert := &x509.Certificate{
		SerialNumber:          big.NewInt(newCertSerial),
		Subject:               csr.Subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		DNSNames:              csr.DNSNames,
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		EmailAddresses:        csr.EmailAddresses,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientCert, ca, csr.PublicKey, caKey)
	if err != nil {
		return nil, err
	}

	clientCertPEM := new(bytes.Buffer)
	pem.Encode(clientCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertDER,
	})

	return clientCertPEM.Bytes(), nil
}

func NewMockClientCert(testContext *hostProviderTestContext) (err error) {
	clientKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	clientKeyPEM := new(bytes.Buffer)
	pem.Encode(clientKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(clientKey),
	})

	if err := testContext.Add(fileIdClientKey, clientKeyPEM.Bytes()); err != nil {
		return err
	}

	clientCert := &x509.Certificate{
		SerialNumber: big.NewInt(firstCertSerial),
		Subject: pkix.Name{
			Organization:  []string{"trivago"},
			Country:       []string{"DE"},
			Province:      []string{""},
			Locality:      []string{"Duesseldorf"},
			StreetAddress: []string{"Kesselstraße 5-7"},
			PostalCode:    []string{"40221"},
		},
		IPAddresses:    []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		EmailAddresses: []string{"test@test.com"},
		NotBefore:      time.Now(),
		NotAfter:       time.Now().Add(time.Hour),
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		KeyUsage:       x509.KeyUsageDigitalSignature,
	}

	clientCertDER, err := x509.CreateCertificate(rand.Reader, clientCert, testContext.ca, &clientKey.PublicKey, testContext.caKey)
	if err != nil {
		return err
	}

	clientCertPEM := new(bytes.Buffer)
	pem.Encode(clientCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: clientCertDER,
	})

	return testContext.Add(fileIdClientCert, clientCertPEM.Bytes())
}

func TestHostTokenProviderRenewFileHandling(t *testing.T) {
	assert := assert.New(t)
	files := &hostProviderTestContext{
		path: make(map[string]string),
	}
	defer files.Clean()

	srv, err := NewMockIdentityServer(files)
	assert.NoError(err)
	defer srv.Close()

	err = NewMockClientCert(files)
	assert.NoError(err)

	// Make sure we've passed the minimum lifetime
	time.Sleep(2 * time.Second)

	provider, err := NewHostTokenProvider(
		"test",
		srv.URL,
		files.path[fileIdCACert],
		files.path[fileIdClientCert],
		files.path[fileIdClientKey],
		time.Minute*5,
		time.Hour-time.Second)

	assert.NoError(err)
	assert.NotNil(provider)

	certLinkInfo, err := os.Lstat(files.path[fileIdClientCert])
	assert.NoError(err)
	assert.NotNil(certLinkInfo)
	assert.Equal(os.ModeSymlink, certLinkInfo.Mode()&os.ModeSymlink)

	keyLinkInfo, err := os.Lstat(files.path[fileIdClientKey])
	assert.NoError(err)
	assert.NotNil(keyLinkInfo)
	assert.Equal(os.ModeSymlink, keyLinkInfo.Mode()&os.ModeSymlink)

	certFileName, err := os.Readlink(files.path[fileIdClientCert])
	assert.NoError(err)

	keyFileName, err := os.Readlink(files.path[fileIdClientKey])
	assert.NoError(err)

	// Check if the new cert is being used
	// The mock server will return the new cert serial number
	identity := provider.GetIdentityForIP(context.Background(), "127.0.0.1")
	assert.Equal(strconv.Itoa(newCertSerial), identity.GetBoundGSA())

	// Make sure we've passed the minimum lifetime again
	time.Sleep(2 * time.Second)

	provider.TryRefreshCertificate()

	newCertFileName, err := os.Readlink(files.path[fileIdClientCert])
	assert.NoError(err)
	assert.NotEqual(certFileName, newCertFileName)

	newKeyFileName, err := os.Readlink(files.path[fileIdClientKey])
	assert.NoError(err)
	assert.NotEqual(keyFileName, newKeyFileName)
}

func TestHostTokenProviderRenew(t *testing.T) {
	assert := assert.New(t)
	files := &hostProviderTestContext{
		path: make(map[string]string),
	}
	defer files.Clean()

	srv, err := NewMockIdentityServer(files)
	assert.NoError(err)
	defer srv.Close()

	err = NewMockClientCert(files)
	assert.NoError(err)

	provider, err := NewHostTokenProvider(
		"test",
		srv.URL,
		files.path[fileIdCACert],
		files.path[fileIdClientCert],
		files.path[fileIdClientKey],
		time.Minute,
		time.Hour-time.Second)

	assert.NoError(err)
	assert.NotNil(provider)

	// Check if the new cert is being used
	// The mock server will return the new cert serial number
	identity := provider.GetIdentityForIP(context.Background(), "127.0.0.1")
	assert.Equal(strconv.Itoa(firstCertSerial), identity.GetBoundGSA())

	// Make sure we clear the cache so the query happens again
	provider.ClearIdentityCache()

	// Make sure we've passed the minimum lifetime
	time.Sleep(2 * time.Second)
	provider.TryRefreshCertificate()

	// Check if the new cert is being used
	// The mock server will return the new cert serial number
	identity = provider.GetIdentityForIP(context.Background(), "127.0.0.1")
	assert.Equal(strconv.Itoa(newCertSerial), identity.GetBoundGSA())
}

func TestHostTokenProviderAutoRenew(t *testing.T) {
	assert := assert.New(t)
	files := &hostProviderTestContext{
		path: make(map[string]string),
	}
	defer files.Clean()

	srv, err := NewMockIdentityServer(files)
	assert.NoError(err)
	defer srv.Close()

	err = NewMockClientCert(files)
	assert.NoError(err)

	provider, err := NewHostTokenProvider(
		"test",
		srv.URL,
		files.path[fileIdCACert],
		files.path[fileIdClientCert],
		files.path[fileIdClientKey],
		time.Second*2,
		time.Hour-time.Second)

	assert.NoError(err)
	assert.NotNil(provider)

	// Check if the new cert is being used
	// The mock server will return the new cert serial number
	identity := provider.GetIdentityForIP(context.Background(), "127.0.0.1")
	assert.Equal(strconv.Itoa(firstCertSerial), identity.GetBoundGSA())

	// Make sure we clear the cache so the query happens again
	provider.ClearIdentityCache()

	// Make sure we've passed the minimum lifetime and allowed the auto-renewal to happen
	time.Sleep(3 * time.Second)

	// Check if the new cert is being used
	// The mock server will return the new cert serial number
	identity = provider.GetIdentityForIP(context.Background(), "127.0.0.1")
	assert.Equal(strconv.Itoa(newCertSerial), identity.GetBoundGSA())
}
