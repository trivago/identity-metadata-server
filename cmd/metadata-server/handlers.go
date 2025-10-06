package main

import (
	"identity-metadata-server/internal/shared"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	subPathMatch *regexp.Regexp
)

func init() {
	subPathMatch = regexp.MustCompile("^[^/]+[/]{0,1}$")
}

// HandleOk returns a 200 status code.
func HandleOk(c *gin.Context) {
	shared.DebugEndpoint(c)
	c.Header("Metadata-Flavor", "Google")
	c.Status(http.StatusOK)
}

// HandleNotFound returns a 404 status code.
func HandleNotFound(c *gin.Context) {
	shared.DebugEndpoint(c)
	c.Status(http.StatusNotFound)
}

// HandleListEndpoints returns a list of endpoints for the given path.
func HandleListEndpoints(c *gin.Context) {
	if !isValidMetadataRequest(c) {
		c.Status(http.StatusBadRequest)
		return
	}

	// TODO: We don't support recursive=true

	path := c.Request.URL.Path
	endpointList := ""

	// Path must end with a slash
	if path[len(path)-1] != '/' {
		path += "/"
	}

	// Get all keys starting with the given path
	for key := range endpoints {
		if key == path {
			continue
		}

		if strings.HasPrefix(key, path) {
			suffix := key[len(path):]
			// We only want to return the next level of endpoints
			if subPathMatch.MatchString(suffix) {
				endpointList += suffix + "\n"
			}
		}
	}

	c.Header("Metadata-Flavor", "Google")
	c.String(http.StatusOK, endpointList)
}
