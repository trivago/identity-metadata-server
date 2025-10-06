package main

import (
	"identity-metadata-server/internal/shared"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// HandleGetAccessToken handles an access token request.
// The tokens returned by this endpoint are GCP-native access tokens.
func HandleGetAccessToken(c *gin.Context) {
	// The call is explained here
	// https://cloud.google.com/compute/docs/access/authenticate-workloads#applications

	if !isValidMetadataRequest(c) {
		c.Status(http.StatusBadRequest)
		return
	}

	// The scope of the token is optional and defaults to the cloud-platform scope.
	// The latter includes all Google Cloud services.
	scopes := []string{shared.DefaultScope}
	if scopeArg := c.Query("scopes"); len(scopeArg) > 0 {
		scopes = strings.Split(scopeArg, ",")
	}

	// The main (first) audience is fixed as at has to be set to the workload
	// identity provider. Additional audiences can be set by the request.
	additionalAudiences := []string{}
	audience := c.Query("audience")
	if len(audience) > 0 {
		additionalAudiences = append(additionalAudiences, audience)
	}

	// Get the Kubernetes Service Account (KSA) to request the token for.
	// We explicitly use RemotIP here over ClientIP, as we want to use the
	// "direct IP", not one that might have been set through a http header.
	// This also means that proxied requests won't work here by design.
	srcIdentity := tokenProvider.GetIdentityForIP(c.Request.Context(), c.ClientIP())

	// Get the google service account (GSA) to authenticate as.
	gsa := c.Param("serviceAccount")
	if len(gsa) == 0 || strings.ToLower(gsa) == "default" {
		gsa = srcIdentity.GetBoundGSA()
	}

	tokenID := NewLookupWithScopeAndAudience(TokenTypeAccess, srcIdentity, scopes, additionalAudiences)
	cachedToken := knownTokens.Get(tokenID)

	if cachedToken == nil {
		// The documentation is a bit patchy here, so we don't know if we can
		// actually override the token lifetime through a request.
		// TODO: Reverse-engineering is required here. We need to find a
		// call that sets the token lifetime and see which parameter is
		// being used.
		tokenLifeTime := AccessTokenLifetime

		trt, err := tokenProvider.GetTokenRequestToken(c.Request.Context(), srcIdentity, tokenLifeTime, scopes, additionalAudiences)
		if trt == nil {
			shared.HttpError(c, http.StatusInternalServerError, err)
			return
		}

		// Get the token for the given parameters.
		accessToken, err := tokenProvider.GetAccessToken(c.Request.Context(), *trt, tokenLifeTime, scopes, gsa)
		if accessToken == nil {
			shared.HttpError(c, http.StatusInternalServerError, err)
			return
		}

		cachedToken = knownTokens.StoreUntil(tokenID, accessToken.AccessToken, accessToken.ExpireTime)
	}

	// The format is explained in the documentation.
	// https://cloud.google.com/compute/docs/access/authenticate-workloads#applications
	// The response format is identical to the one used by the STS endpoint,
	// which might be by design, but is not documented.
	response := shared.TokenExchangeResponse{
		AccessToken: cachedToken.token,
		ExpiresIn:   int(time.Until(cachedToken.expires).Seconds()),
		TokenType:   "Bearer",
	}

	c.Header("Metadata-Flavor", "Google")
	c.JSON(http.StatusOK, response)
}

// HandleGetIdentityToken handles an identity token request.
// These tokens are JWT tokens signed by the workload identity provider.
func HandleGetIdentityToken(c *gin.Context) {
	// The call is explained here
	// https://cloud.google.com/compute/docs/instances/verifying-instance-identity

	if !isValidMetadataRequest(c) {
		c.Status(http.StatusBadRequest)
		return
	}

	// The audience _has_ to be set for this request
	audience := c.Query("audience")
	if len(audience) == 0 {
		c.String(http.StatusBadRequest, "audience parameter is required")
		return
	}

	// Get the Kubernetes Service Account (KSA) to request the token for.
	// We explicitly use RemotIP here over ClientIP, as we want to use the
	// "direct IP", not one that might have been set through a http header.
	// This also means that proxied requests won't work here by design.
	srcIdentity := tokenProvider.GetIdentityForIP(c.Request.Context(), c.ClientIP())

	// Get the google service account (GSA) to authenticate as.
	gsa := c.Param("serviceAccount")
	if len(gsa) == 0 || strings.ToLower(gsa) == "default" {
		gsa = srcIdentity.GetBoundGSA()
	}

	tokenID := NewLookupWithAudience(TokenTypeIdentity, srcIdentity, audience)
	cachedToken := knownTokens.Get(tokenID)

	if cachedToken == nil {
		trt, err := tokenProvider.GetTokenRequestToken(c.Request.Context(), srcIdentity, IdentityTokenLifetime, []string{shared.IdentityTokenScope}, []string{audience})
		if trt == nil {
			shared.HttpError(c, http.StatusInternalServerError, err)
			return
		}

		// Get the token for the given parameters.
		idToken, err := tokenProvider.GetIdentityToken(c.Request.Context(), *trt, gsa, audience)
		if idToken == nil {
			shared.HttpError(c, http.StatusInternalServerError, err)
			return
		}

		cachedToken = knownTokens.StoreFor(tokenID, idToken.Token, IdentityTokenLifetime)
	}

	// The token returned by this endpoint is a plain, signed JWT token.
	// https://cloud.google.com/compute/docs/instances/verifying-instance-identity#token_format

	c.Header("Metadata-Flavor", "Google")
	c.String(http.StatusOK, cachedToken.token)
}
