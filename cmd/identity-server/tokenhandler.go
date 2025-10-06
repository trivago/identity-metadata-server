package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"identity-metadata-server/internal/shared"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type TokenRequest struct {
	Audiences []string `json:"audiences"`
	Lifetime  string   `json:"lifetime,omitempty"`
}

// HandleTokenRequest generates a new token for the given audience.
func HandleTokenRequest(c *gin.Context, crl *CertificateRevocationList) {
	// Check if we can get a client from the context
	client, err := NewClientFromContext(c, crl)
	if err != nil {
		log.Error().Msg("Failed to get client from context")
		shared.HttpError(c, http.StatusInternalServerError, err)
		return
	}

	// Parse the token request data and make sure it's not malformed
	tokenRequestData := TokenRequest{
		Lifetime: "10m",
	}
	if err := c.BindJSON(&tokenRequestData); err != nil {
		log.Error().Err(err).Msg("Failed to parse token request")
		shared.HttpError(c, http.StatusBadRequest, err)
		return
	}

	if len(tokenRequestData.Audiences) == 0 {
		log.Error().Msg("Blocked token request with empty audience")
		shared.HttpErrorString(c, http.StatusBadRequest, "Audience must not be empty")
		return
	}

	lifetime, err := time.ParseDuration(tokenRequestData.Lifetime)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse token lifetime")
		shared.HttpErrorString(c, http.StatusBadRequest, "Invalid token lifetime")
		return
	}

	// Generate a JWT token
	oidcToken, err := generateOIDCToken(client.Identity, client.Host, tokenRequestData.Audiences, lifetime)
	if err != nil {
		log.Error().Err(err).Msg("Failed to sign jwt token")
		shared.HttpError(c, http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, oidcToken)
}

// generateOIDCToken generates a new OIDC token for the given serviceAccount and hostname.
// It uses the provided audiences and lifetime to create the token.
// The token is signed using the private key of the identity server.
func generateOIDCToken(serviceAccount, hostname string, audiences []string, lifetime time.Duration) (string, error) {
	now := time.Now()

	// The JWTID is a unique identifier for the token. It is used to prevent replay attacks.
	// We create a hash of the service account and a random number to make it unique.
	// If the random number cannot be generated, we use the current time instead.
	// This way ID is cryptographically secure and unique for each token request.
	// See https://datatracker.ietf.org/doc/html/rfc7519#section-4.1.7
	jwtID := sha256.New()
	jwtID.Write([]byte(serviceAccount))

	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err == nil {
		jwtID.Write(randomBytes)
	} else {
		jwtID.Write([]byte(now.String()))
	}

	claims := CustomClaims{
		NodeClaims: NodeClaims{
			Identity: serviceAccount,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    viper.GetString("server.issuer"),
			Subject:   hostname,
			Audience:  audiences,
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ID:        hex.EncodeToString(jwtID.Sum(nil)),
		},
	}

	return buildAndSignJWT(claims)
}
