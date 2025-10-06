package shared

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"

	jsoniter "github.com/json-iterator/go"
)

var (
	// KnownRootCAs is a pool of known root CAs.
	// If this pool is nil, the system pool will be used.
	KnownRootCAs *x509.CertPool
)

func init() {
	KnownRootCAs, _ = x509.SystemCertPool()
	if KnownRootCAs == nil {
		KnownRootCAs = x509.NewCertPool()
	}
}

// debug logs the incoming request.
func DebugEndpoint(c *gin.Context) {
	data, _ := c.GetRawData()

	log.Info().
		Interface("headers", c.Request.Header).
		Interface("path", c.Request.URL.Path).
		Interface("query", c.Request.URL.Query()).
		Interface("params", c.Params).
		Interface("remote", c.Request.RemoteAddr).
		Msg(string(data))
}

// RegisterCA reads a certificate from disk and registers it as a root CA
// by calling RegisterRootCA.
// This functions is not thread-safe and is meant to be called during
// initialization of the application.
func RegisterRootCAFile(caFilePath string) error {
	caCert, err := os.ReadFile(caFilePath)
	if err != nil {
		return WrapErrorf(err, "failed to read control plane root CA from %s", caFilePath)
	}

	if !RegisterRootCA(caCert) {
		return fmt.Errorf("failed to register control plane root CA")
	}
	return nil
}

// RegisterRootCA registers a new root CA to the KnownRootCAs pool.
// If the KnownRootCAs pool is nil, a new pool will be created. By default,
// the system pool is used as a base.
// This functions is not thread-safe and is meant to be called during
// initialization of the application.
func RegisterRootCA(caCert []byte) bool {
	return KnownRootCAs.AppendCertsFromPEM(caCert)
}

// ReadBodyWithLimit reads the body of the response with a limit on the number of bytes.
// If the content length is greater than maxBytes, it will read only maxBytes bytes.
// If maxBytes is less than or equal to 0, it will read the entire body.
// If the response or body is nil, it returns an error and an empty byte slice.
func ReadBodyWithLimit(rsp *http.Response, maxBytes int) ([]byte, error) {
	if rsp == nil || rsp.Body == nil {
		return []byte{}, errors.New("response or body is nil")
	}

	if rsp.ContentLength <= 0 {
		return []byte{}, errors.New("content length is zero or negative")
	}

	if maxBytes > 0 && rsp.ContentLength > int64(maxBytes) {
		content := make([]byte, maxBytes)
		n, err := io.ReadFull(rsp.Body, content)
		return content[:n], err
	}

	return io.ReadAll(rsp.Body)
}

// HttpPOST sends a POST request to the given address with the given body and headers.
func HttpPOST(address string, body []byte, header map[string]string, cert *tls.Certificate, ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST",
		address,
		bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	client := newHttpClient(cert)
	return client.Do(req)
}

// HttpPOST sends a POST request to the given address with the given body and headers.
func HttpGET(address string, body []byte, header map[string]string, cert *tls.Certificate, ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET",
		address,
		bytes.NewReader(body))

	if err != nil {
		return nil, err
	}

	for key, value := range header {
		req.Header.Set(key, value)
	}

	client := newHttpClient(cert)
	return client.Do(req)
}

// HttpGETJson sends a GET request to the given address with the given body and headers.
// It returns the response body as a JSON decoded object of type T.
// In case of a non-200 status code, it reads up to 32KB of the response body and includes it in the error message.
// If the response body is empty, it returns an error with a textual representation of the status code.
func HttpGETJson[T any](address string, body []byte, header map[string]string, cert *tls.Certificate, ctx context.Context) (*T, error) {
	if header == nil {
		header = make(map[string]string)
	}
	if _, ok := header["Accept"]; !ok {
		header["Accept"] = "application/json"
	}

	resp, err := HttpGET(address, body, header, cert, ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		rspData, err := ReadBodyWithLimit(resp, 32*1024)
		if len(rspData) == 0 {
			err = errors.Join(err, errors.New(http.StatusText(resp.StatusCode)))
		} else {
			err = errors.Join(err, errors.New(string(rspData)))
		}
		return nil, WrapErrorWithStatus(err, resp.StatusCode)
	}

	var result T
	if err := jsoniter.ConfigFastest.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// HttpPOSTJson sends a POST request to the given address with the given body and headers.
// It returns the response body as a JSON decoded object of type T.
// In case of a non-200 status code, it reads up to 32KB of the response body and includes it in the error message.
func HttpPOSTJson[T any](address string, body []byte, header map[string]string, cert *tls.Certificate, ctx context.Context) (*T, error) {
	if header == nil {
		header = make(map[string]string)
	}
	if _, ok := header["Accept"]; !ok {
		header["Accept"] = "application/json"
	}

	resp, err := HttpPOST(address, body, header, cert, ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		rspData, _ := ReadBodyWithLimit(resp, 32*1024)
		if len(rspData) == 0 {
			err = errors.New(http.StatusText(resp.StatusCode))
		} else {
			err = errors.New(string(rspData))
		}
		return nil, WrapErrorWithStatus(err, resp.StatusCode)
	}

	var result T
	if err := jsoniter.ConfigFastest.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// newHttpClient creates a new http2 client with the given certificate or
// the default client if no certificate is given.
func newHttpClient(cert *tls.Certificate) *http.Client {
	tlsConfig := &tls.Config{
		RootCAs: KnownRootCAs,
	}

	if cert != nil {
		tlsConfig.Certificates = []tls.Certificate{*cert}
		return &http.Client{
			Transport: &http2.Transport{
				TLSClientConfig: tlsConfig,
			},
		}
	}

	// We use HTTP1 transport if no certificate is provided, as
	// some API endpoints might not support unencrypted HTTP2.
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
}
