package main

import (
	"identity-metadata-server/internal/shared"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceAccountInfo struct {
	Aliases []string `json:"aliases"`
	Email   string   `json:"email"`
	Scopes  []string `json:"scopes"`
}

func HandleGetServiceAccountScopes(c *gin.Context) {
	if !isValidMetadataRequest(c) {
		c.Status(http.StatusBadRequest)
		return
	}

	c.Header("Metadata-Flavor", "Google")
	c.String(http.StatusOK, shared.DefaultScope+"\n") // Always return the default scope
}

// HandleGetDefaultServiceAccount returns the default service account
// as defined by the configuration.
func HandleGetDefaultServiceAccount(c *gin.Context) {
	if !isValidMetadataRequest(c) {
		c.Status(http.StatusBadRequest)
		return
	}

	srcIdentity := tokenProvider.GetIdentityForIP(c.Request.Context(), c.ClientIP())

	c.Header("Metadata-Flavor", "Google")
	c.String(http.StatusOK, srcIdentity.GetBoundGSA())
}

// HandleGetServiceAccounts returns the list of available service accounts
func HandleGetServiceAccounts(c *gin.Context) {
	if !isValidMetadataRequest(c) {
		c.Status(http.StatusBadRequest)
		return
	}

	srcIdentity := tokenProvider.GetIdentityForIP(c.Request.Context(), c.ClientIP())
	serviceAccountEmail := srcIdentity.GetBoundGSA()

	// If `recursive` is set, JSON is returned instead of plain text.
	// The expected fields per object are explained here:
	// https://cloud.google.com/compute/docs/metadata/predefined-metadata-keys#instance-metadata
	//
	// TODO: The "identity" and "token" fields are not yet implemented, as it's not quite clear
	// how they should be populated, and when exactly they are expected in this context.

	if c.Query("recursive") == "true" {
		response := map[string]ServiceAccountInfo{
			serviceAccountEmail: {
				Email:  serviceAccountEmail,
				Scopes: []string{shared.DefaultScope},
				Aliases: []string{
					"default",
				},
			},
			"default": {
				Email:  serviceAccountEmail,
				Scopes: []string{shared.DefaultScope},
				Aliases: []string{
					"default",
				},
			},
		}

		c.Header("Metadata-Flavor", "Google")
		c.JSON(http.StatusOK, response)
		return
	}

	// Plain text response with one service account per line
	// The first line is always the main service account

	response := serviceAccountEmail + "/\n"
	response += "default/\n"

	c.Header("Metadata-Flavor", "Google")
	c.String(http.StatusOK, response) // Force use of the default service account
}

// HandleGetServiceAccountInfo returns information about a single service account
func HandleGetServiceAccountInfo(c *gin.Context) {
	if !isValidMetadataRequest(c) {
		c.Status(http.StatusBadRequest)
		return
	}

	info := ServiceAccountInfo{
		Email:  c.Param("serviceAccount"),
		Scopes: []string{shared.DefaultScope},
	}

	srcIdentity := tokenProvider.GetIdentityForIP(c.Request.Context(), c.ClientIP())

	switch info.Email {
	case "default":
		info.Email = srcIdentity.GetBoundGSA()
		info.Aliases = []string{"default"}

	case srcIdentity.GetBoundGSA():
		info.Aliases = []string{"default"}
	}

	c.Header("Metadata-Flavor", "Google")
	c.JSON(http.StatusOK, info)
}
