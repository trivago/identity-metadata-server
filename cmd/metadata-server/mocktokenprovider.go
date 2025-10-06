package main

import (
	"context"
	"hash"
	"identity-metadata-server/internal/shared"
	"identity-metadata-server/internal/tokenprovider"
	"strings"
	"time"

	"github.com/cespare/xxhash"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

// MockTokenProvider is a mock implementation of the TokenProvider interface
// to be used in tests.
type MockTokenProvider struct{}

// MockSourceIdentity is a mock implementation of the SourceIdentity interface
type MockSourceIdentity struct {
	Name     string `json:"name"`
	BoundGSA string `json:"boundGSA"`
}

// MockToken is used to test if values are passed correctly between token functions
type MockToken struct {
	Identity  tokenprovider.SourceIdentity `json:"identity"`
	Scopes    []string                     `json:"scopes"`
	Audiences []string                     `json:"audiences"`
}

func NewMockTokenProvider() *MockTokenProvider {
	return &MockTokenProvider{}
}

// GetIdentityForIP returns a fake kubernetesServiceAccountInfo object
func (tp *MockTokenProvider) GetIdentityForIP(ctx context.Context, ip string) tokenprovider.SourceIdentity {
	return MockSourceIdentity{
		Name:     ip,
		BoundGSA: "test@gcp.project",
	}
}

// GetTokenRequestToken returns a fake TokenExhcangeResponse object.
// The returned token is a MockToken with the source identity, scope and
// additional audiences.
func (tp *MockTokenProvider) GetTokenRequestToken(ctx context.Context, srcIdentity tokenprovider.SourceIdentity, lifetime time.Duration, scopes, additionalAudiences []string) (*shared.TokenExchangeResponse, error) {
	token := MockToken{
		Identity:  srcIdentity,
		Scopes:    scopes,
		Audiences: additionalAudiences,
	}

	fakeToken, _ := jsoniter.MarshalToString(token)
	return &shared.TokenExchangeResponse{
		AccessToken: fakeToken,
		ExpiresIn:   int(lifetime.Seconds()),
		TokenType:   "Bearer",
	}, nil
}

// GetIdentityToken returns a fake shared.IAMIdentityTokenResponse object.
// The returned tolen is a MockToken with the name and scope of the source
// identity but the requested GSA and audience.
func (tp *MockTokenProvider) GetIdentityToken(ctx context.Context, tokenRequestToken shared.TokenExchangeResponse, gsa string, audience string) (*shared.IAMIdentityTokenResponse, error) {
	trtToken := NewMockToken()
	err := jsoniter.UnmarshalFromString(tokenRequestToken.AccessToken, &trtToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal token request token")
		return nil, err
	}

	token := MockToken{
		Identity: MockSourceIdentity{
			Name:     trtToken.Identity.(*MockSourceIdentity).Name,
			BoundGSA: gsa,
		},
		Scopes:    trtToken.Scopes,
		Audiences: []string{audience},
	}
	fakeToken, _ := jsoniter.MarshalToString(token)

	return &shared.IAMIdentityTokenResponse{
		Token: fakeToken,
	}, nil
}

// GetAccessToken returns a fake shared.IAMAccessTokenResponse object.
// The returned token is a MockToken with the name and audience of the source
// identity but the requested GSA and scope.
func (tp *MockTokenProvider) GetAccessToken(ctx context.Context, tokenRequestToken shared.TokenExchangeResponse, lifetime time.Duration, scopes []string, gsa string) (*shared.IAMAccessTokenResponse, error) {
	trtToken := NewMockToken()
	err := jsoniter.UnmarshalFromString(tokenRequestToken.AccessToken, &trtToken)
	if err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal token request token")
		return nil, err
	}

	token := MockToken{
		Identity: MockSourceIdentity{
			Name:     trtToken.Identity.(*MockSourceIdentity).Name,
			BoundGSA: gsa,
		},
		Scopes:    scopes,
		Audiences: trtToken.Audiences,
	}
	fakeToken, _ := jsoniter.MarshalToString(token)

	return &shared.IAMAccessTokenResponse{
		AccessToken: fakeToken,
		ExpireTime:  time.Now().Add(lifetime).Format(tokenTimeFormat),
	}, nil
}

// NewMockToken returns a new MockToken object with the identity set to a
// MockSourceIdentity object. This prevents nil pointer dereference errors
// when unmarshalling the token.
func NewMockToken() MockToken {
	return MockToken{
		Identity: &MockSourceIdentity{},
	}
}

// Hash returns a hash of the mock identity
func (id MockSourceIdentity) Hash() hash.Hash64 {
	hash := xxhash.New()
	hash.Write([]byte(strings.Join([]string{id.BoundGSA, id.Name}, ";")))
	return hash
}

// GetBoundGSA returns the bound GSA of the mock identity
func (id MockSourceIdentity) GetBoundGSA() string {
	return id.BoundGSA
}

func (h MockSourceIdentity) Equal(other tokenprovider.SourceIdentity) bool {
	h2, isSameType := other.(MockSourceIdentity)
	return isSameType &&
		h.Name == h2.Name &&
		h.BoundGSA == h2.BoundGSA
}
