package main

import (
	"context"
	"identity-metadata-server/internal/shared"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type mockServer struct {
	init   sync.Once
	router *gin.Engine
}

var (
	TestServer mockServer
)

func (m *mockServer) GetRouter() *gin.Engine {
	m.init.Do(func() {
		tokenProvider = NewMockTokenProvider()
		knownTokens = NewTokenCache(0, 0)
		m.router = gin.Default()
		initGinEndpoints(m.router)
		initConfigDefaults()
	})
	return m.router
}

func TestHandleHttpRoot(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
}

func TestHandleUniverseDomain(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/computeMetadata/v1/universe/universe-domain", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
	assert.Equal("googleapis.com", w.Body.String())
}

func TestHandleGetProjectId(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/computeMetadata/v1/project/project-id", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
	assert.Equal(viper.Get("projectId"), w.Body.String())
}

func TestHandleGetProjectNumber(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/computeMetadata/v1/project/numeric-project-id", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
	assert.Equal(viper.Get("projectNumber"), w.Body.String())
}

func TestHandleGetDefaultServiceAccount(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/computeMetadata/v1/instance/service-accounts/default/email", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
	assert.Equal("test@gcp.project", w.Body.String())
}

func TestHandleGetServiceAccounts(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/computeMetadata/v1/instance/service-accounts/", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
	assert.Equal("test@gcp.project/\ndefault/\n", w.Body.String())
}

func TestHandleGetServiceAccountInfo(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/computeMetadata/v1/instance/service-accounts/test@gcp.project", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected := ServiceAccountInfo{
		Email:   "test@gcp.project",
		Scopes:  []string{shared.DefaultScope},
		Aliases: []string{"default"},
	}
	expectedJson, _ := jsoniter.MarshalToString(expected)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
	assert.Equal(expectedJson, w.Body.String())

	// Check default SA retreival

	req, _ = http.NewRequest("GET", "/computeMetadata/v1/instance/service-accounts/default", strings.NewReader(``))
	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
	assert.Equal(expectedJson, w.Body.String())

	// Check other SA retreival

	req, _ = http.NewRequest("GET", "/computeMetadata/v1/instance/service-accounts/foobar", strings.NewReader(``))
	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expected.Email = "foobar"
	expected.Aliases = nil
	expectedJson, _ = jsoniter.MarshalToString(expected)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
	assert.Equal(expectedJson, w.Body.String())
}

func TestHandleGetAccessToken(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/computeMetadata/v1/instance/service-accounts/default/token", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expectedToken := MockToken{
		Identity:  tokenProvider.GetIdentityForIP(context.Background(), ""), // No IP for httptest
		Scopes:    []string{shared.DefaultScope},
		Audiences: []string{},
	}
	expectedTokenJson, _ := jsoniter.MarshalToString(expectedToken)

	expected := shared.TokenExchangeResponse{
		AccessToken: expectedTokenJson,
		ExpiresIn:   int(AccessTokenLifetime.Seconds()),
		TokenType:   "Bearer",
	}

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))

	parsedBody := shared.TokenExchangeResponse{}
	err := jsoniter.UnmarshalFromString(w.Body.String(), &parsedBody)
	assert.NoError(err)

	assert.Equal(expected.AccessToken, parsedBody.AccessToken)
	assert.Equal(expected.TokenType, parsedBody.TokenType)
	assert.LessOrEqual(parsedBody.ExpiresIn, expected.ExpiresIn)
}

func TestHandleGetAccessTokenArgs(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/computeMetadata/v1/instance/service-accounts/default/token?scopes=test1,test2&audience=test3", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	expectedToken := MockToken{
		Identity:  tokenProvider.GetIdentityForIP(context.Background(), ""), // No IP for httptest
		Scopes:    []string{"test1", "test2"},
		Audiences: []string{"test3"},
	}
	expectedTokenJson, _ := jsoniter.MarshalToString(expectedToken)

	expected := shared.TokenExchangeResponse{
		AccessToken: expectedTokenJson,
		ExpiresIn:   int(AccessTokenLifetime.Seconds()),
		TokenType:   "Bearer",
	}

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))

	parsedBody := shared.TokenExchangeResponse{}
	err := jsoniter.UnmarshalFromString(w.Body.String(), &parsedBody)
	assert.NoError(err)

	assert.Equal(expected.AccessToken, parsedBody.AccessToken)
	assert.Equal(expected.TokenType, parsedBody.TokenType)
	assert.LessOrEqual(parsedBody.ExpiresIn, expected.ExpiresIn)
}

func TestHandleGeIdentityToken(t *testing.T) {
	assert := assert.New(t)
	router := TestServer.GetRouter()

	req, _ := http.NewRequest("GET", "/computeMetadata/v1/instance/service-accounts/default/identity", strings.NewReader(``))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(http.StatusBadRequest, w.Code)

	req, _ = http.NewRequest("GET", "/computeMetadata/v1/instance/service-accounts/default/identity?audience=test", strings.NewReader(``))
	req.Header.Set("Metadata-Flavor", "Google")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)

	expectedToken := MockToken{
		Identity:  tokenProvider.GetIdentityForIP(context.Background(), ""), // No IP for httptest
		Scopes:    []string{shared.IdentityTokenScope},
		Audiences: []string{"test"},
	}
	expectedTokenJson, _ := jsoniter.MarshalToString(expectedToken)

	assert.Equal(http.StatusOK, w.Code)
	assert.Equal("Google", w.Header().Get("Metadata-Flavor"))
	assert.Equal(expectedTokenJson, w.Body.String())
}
