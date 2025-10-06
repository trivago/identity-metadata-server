package tokenprovider

import (
	"context"
	"hash"
	"identity-metadata-server/internal/shared"
	"time"
)

// SourceIdentity is an interface that represents the identity of an object
// that initiated a token request.
type SourceIdentity interface {
	Hash() hash.Hash64
	GetBoundGSA() string
	Equal(other SourceIdentity) bool
}

// TokenRequestProvider is an interface that is used to implement the first step
// of the token exchange process. Functions resolve around getting the identity
// and request token to be used in with a token exchange provider.
type TokenRequestProvider interface {
	GetIdentityForIP(ctx context.Context, ip string) SourceIdentity
	GetTokenRequestToken(ctx context.Context, srcIdentity SourceIdentity, lifetime time.Duration, scopes, additionalAudiences []string) (*shared.TokenExchangeResponse, error)
}

// TokenExchangeProvider is an interface that is used to implement the final stept
// of the token exchange process.
type TokenExchangeProvider interface {
	GetIdentityToken(ctx context.Context, tokenRequestToken shared.TokenExchangeResponse, gsa string, audience string) (*shared.IAMIdentityTokenResponse, error)
	GetAccessToken(ctx context.Context, tokenRequestToken shared.TokenExchangeResponse, lifetime time.Duration, scopes []string, gsa string) (*shared.IAMAccessTokenResponse, error)
}

// TokenProvider is an interface that is implementing to sides of the token exchange
// process. It can be used to get the identity of the source of the token request
// and to get the actual tokens.
type TokenProvider interface {
	TokenRequestProvider
	TokenExchangeProvider
}
