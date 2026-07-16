package main

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newProtocolTestServer starts a cleartext listener with HTTP/1.1 and h2c
// enabled, matching production metadata-server configuration.
func newProtocolTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	gin.SetMode(gin.TestMode)
	initConfigDefaults()
	router := gin.New()
	initGinEndpoints(router)

	server := httptest.NewUnstartedServer(router)
	configureHTTPServer(server.Config, 620*time.Second)
	server.Start()

	t.Cleanup(server.Close)
	return server
}

func TestMetadataServerProtocols(t *testing.T) {
	t.Parallel()

	server := newProtocolTestServer(t)

	tests := []struct {
		name               string
		client             *http.Client
		wantProtoMajor     int
		wantMetadataFlavor string
	}{
		{
			name:               "HTTP/1.1",
			client:             server.Client(),
			wantProtoMajor:     1,
			wantMetadataFlavor: "Google",
		},
		{
			name: "h2c prior knowledge",
			client: &http.Client{
				Transport: unencryptedHTTP2Transport(),
			},
			wantProtoMajor:     2,
			wantMetadataFlavor: "Google",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest(http.MethodGet, server.URL+"/", nil)
			require.NoError(t, err)

			resp, err := tt.client.Do(req)
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tt.wantProtoMajor, resp.ProtoMajor)
			assert.Equal(t, tt.wantMetadataFlavor, resp.Header.Get("Metadata-Flavor"))
			assert.Empty(t, body)
		})
	}
}

// unencryptedHTTP2Transport returns a transport that uses HTTP/2 with prior
// knowledge for http:// URLs.
func unencryptedHTTP2Transport() *http.Transport {
	transport := &http.Transport{
		DialContext: (&net.Dialer{}).DialContext,
	}
	transport.Protocols = new(http.Protocols)
	transport.Protocols.SetUnencryptedHTTP2(true)
	return transport
}
