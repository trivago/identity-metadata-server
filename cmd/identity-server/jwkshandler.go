package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HandleJWKSRequest returns the JWKS as a JSON response.
// This endpoint is used by clients to fetch the server's public key.
func HandleJWKSRequest(c *gin.Context) {
	// This endpoint is public, no need to check the client certificate
	c.JSON(http.StatusOK, jwks)
}
