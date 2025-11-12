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
	"strconv"
	"time"

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
func HttpPOST(address string, body []byte, header map[string]string, cert *tls.Certificate, retries int, ctx context.Context) (*http.Response, error) {
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
	return DoWithRetry(retries, client, req)
}

// HttpPOST sends a POST request to the given address with the given body and headers.
func HttpGET(address string, body []byte, header map[string]string, cert *tls.Certificate, retries int, ctx context.Context) (*http.Response, error) {
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
	return DoWithRetry(retries, client, req)
}

// HttpGETJson sends a GET request to the given address with the given body and headers.
// It returns the response body as a JSON decoded object of type T.
// In case of a non-200 status code, it reads up to 32KB of the response body and includes it in the error message.
// If the response body is empty, it returns an error with a textual representation of the status code.
func HttpGETJson[T any](address string, body []byte, header map[string]string, cert *tls.Certificate, retries int, ctx context.Context) (*T, error) {
	if header == nil {
		header = make(map[string]string)
	}
	if _, ok := header["Accept"]; !ok {
		header["Accept"] = "application/json"
	}

	resp, err := HttpGET(address, body, header, cert, retries, ctx)
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
func HttpPOSTJson[T any](address string, body []byte, header map[string]string, cert *tls.Certificate, retries int, ctx context.Context) (*T, error) {
	if header == nil {
		header = make(map[string]string)
	}
	if _, ok := header["Accept"]; !ok {
		header["Accept"] = "application/json"
	}

	resp, err := HttpPOST(address, body, header, cert, retries, ctx)
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

// DoWithRetry performs the given HTTP request with retries for specific status codes.
// It retries up to `count` times for status codes that indicate a retry is appropriate,
// such as 429 Too Many Requests. The function waits for a specified duration before each retry,
// which can be influenced by the Retry-After header in the response.
// If the context is cancelled while waiting, it returns an error.
func DoWithRetry(count int, client *http.Client, req *http.Request) (*http.Response, error) {
	// Extract context from request or create a new one
	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	// Perform the request
	rsp, err := client.Do(req)
	if err != nil {
		return rsp, err
	}

	// Return if there's no retry left
	if count <= 0 {
		return rsp, nil
	}

	// We always wait a bit before retrying
	waitDuration := time.Second

	switch rsp.StatusCode {
	// Handle rate limiting. We retry after the given time in the
	// Retry-After header, or after the default waitDuration if not given.
	case http.StatusTooManyRequests:
		if waitDurationStr := rsp.Header.Get("Retry-After"); len(waitDurationStr) > 0 {
			if waitDurationSec, err := strconv.Atoi(waitDurationStr); err == nil && waitDuration > 0 {
				waitDuration = time.Duration(waitDurationSec) * time.Second
			}
		}
		log.Info().Msgf("Received 429 Too Many Requests")

	default:
		// No retry for other status codes
		return rsp, nil
	}

	// Always wait a bit before retrying
	log.Debug().Msgf("Waiting %s before retrying request to %s (remaining retries: %d)...", waitDuration.String(), req.URL.String(), count-1)
	select {
	case <-time.After(waitDuration):
	case <-ctx.Done():
		return nil, WrapErrorf(ctx.Err(), "request cancelled while waiting to retry after 429 Too Many Requests")
	}

	return DoWithRetry(count-1, client, req)
}

// ForceMaxDuration ensures that the context of the given gin.Context
// has at most the given timeout duration.
// If the existing context has a shorter deadline, it is not modified.
// If the existing context has no deadline, a new context with the given timeout is created.
// The cancel function of the new context is scheduled to be called after the timeout duration.
func ForceMaxDuration(timeout time.Duration, ginCtx *gin.Context) {
	parentCtx := ginCtx.Request.Context()
	if deadline, ok := parentCtx.Deadline(); !ok || time.Until(deadline) > timeout {
		newCtx, cancel := context.WithTimeout(parentCtx, timeout)
		defer cancel()
		ginCtx.Request = ginCtx.Request.WithContext(newCtx)
	}
	ginCtx.Next()
}
