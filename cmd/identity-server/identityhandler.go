package main

import (
	"identity-metadata-server/internal/shared"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// HandleIdentityRequest returns the identity of the client.
func HandleIdentityRequest(c *gin.Context, crl *CertificateRevocationList) {
	// Check if we can get a client from the context
	client, err := NewClientFromContext(c, crl)
	if err != nil {
		log.Error().Err(err).Str("clientIP", c.ClientIP()).Msg("Failed to validate client identity")
		shared.HttpError(c, http.StatusInternalServerError, err)
		return
	}

	c.String(http.StatusOK, "%s\n", client.Identity)
}
