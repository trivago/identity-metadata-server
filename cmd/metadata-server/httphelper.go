package main

import (
	"github.com/gin-gonic/gin"
)

// isValidMetadataRequest returns true if the request has the correct metadata
// flavor set.
func isValidMetadataRequest(c *gin.Context) bool {
	return c.Request.Header.Get("Metadata-Flavor") == "Google"
}
