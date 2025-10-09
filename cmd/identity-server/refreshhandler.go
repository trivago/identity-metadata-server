package main

import (
	"identity-metadata-server/internal/shared"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

var (
	lastRefreshGuard = new(sync.Mutex)
	lastRefreshTime  time.Time
)

// HandleRefreshRequest handles a request to refresh the certificate revocation list.
// It rate-limits refreshes to at most once per minute.
func HandleRefreshRequest(c *gin.Context, crl *CertificateRevocationList) {
	lastRefreshGuard.Lock()
	defer lastRefreshGuard.Unlock()

	timeSinceLastRefresh := time.Since(lastRefreshTime)
	if timeSinceLastRefresh < time.Minute {
		remaining := time.Minute - timeSinceLastRefresh
		log.Info().Dur("retry_after", remaining).Msg("CRL refresh request rate-limited")
		c.Header("Retry-After", strconv.Itoa(int(remaining.Seconds())+1))
		c.String(http.StatusTooManyRequests, "Rate limit exceeded")
		return
	}

	err := crl.Update(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to update certificate revocation list")
		shared.HttpError(c, http.StatusInternalServerError, err)
		return
	}

	log.Info().Msg("CRL refresh completed successfully")
	lastRefreshTime = time.Now()
	c.Status(http.StatusOK)
}
