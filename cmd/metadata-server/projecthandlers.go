package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// HandleGetUniverse returns the universe domain, which happens to
// be always googleapis.com.
func HandleGetUniverse(c *gin.Context) {
	c.Header("Metadata-Flavor", "Google")
	c.String(http.StatusOK, "googleapis.com")
}

// HandleGetProjectId returns the project ID as defined by the configuration.
func HandleGetProjectId(c *gin.Context) {
	if !isValidMetadataRequest(c) {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Header("Metadata-Flavor", "Google")
	c.String(http.StatusOK, viper.GetString("projectId"))
}

// HandleGetProjectNumber returns the project number as defined by the configuration.
func HandleGetProjectNumber(c *gin.Context) {
	if !isValidMetadataRequest(c) {
		c.Status(http.StatusBadRequest)
		return
	}
	c.Header("Metadata-Flavor", "Google")
	c.String(http.StatusOK, viper.GetString("projectNumber"))
}
