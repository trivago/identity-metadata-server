package main

import (
	"identity-metadata-server/internal/shared"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

// KnownToken is a token that has been previously requested.
type KnownToken struct {
	token      string
	identifier TokenLookup
	expires    time.Time
}

// TokenCache is a cache for previously requested tokens
type TokenCache struct {
	lock             *sync.Mutex
	data             map[TokenUID]KnownToken
	gcTimer          *time.Timer
	minTokenLifetime time.Duration
	hitMetric        prometheus.Counter
	missMetric       prometheus.Counter
	setMetric        prometheus.Counter
}

// NewTokenCache creates a new token cache with a garbage collection interval.
func NewTokenCache(gcInterval, minLifetime time.Duration) *TokenCache {
	hitMetric := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "metadata_server",
		Subsystem:   "tokencache",
		Name:        "hits_total",
		Help:        "Total number of hits to the token cache.",
		ConstLabels: map[string]string{},
	})
	missMetric := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "metadata_server",
		Subsystem:   "tokencache",
		Name:        "misses_total",
		Help:        "Total number of misses to the token cache.",
		ConstLabels: map[string]string{},
	})
	setMetric := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace:   "metadata_server",
		Subsystem:   "tokencache",
		Name:        "sets_total",
		Help:        "Total number of writes to the token cache.",
		ConstLabels: map[string]string{},
	})

	if err := shared.RegisterCollectorOrUseExisting(&hitMetric); err != nil {
		log.Warn().Err(err).Msg("Failed to register token cache hit metric, metrics will not be available")
	}
	if err := shared.RegisterCollectorOrUseExisting(&missMetric); err != nil {
		log.Warn().Err(err).Msg("Failed to register token cache miss metric, metrics will not be available")
	}
	if err := shared.RegisterCollectorOrUseExisting(&setMetric); err != nil {
		log.Warn().Err(err).Msg("Failed to register token cache set metric, metrics will not be available")
	}

	cache := &TokenCache{
		lock:             &sync.Mutex{},
		data:             make(map[TokenUID]KnownToken),
		minTokenLifetime: minLifetime,
		hitMetric:        hitMetric,
		missMetric:       missMetric,
		setMetric:        setMetric,
	}

	if gcInterval > 0 {
		cache.gcTimer = time.NewTimer(gcInterval)
		go func() {
			for {
				if _, ok := <-cache.gcTimer.C; !ok {
					break
				}
				cache.GC()
				cache.gcTimer.Reset(gcInterval)
			}
		}()
	}

	return cache
}

// StopGC stops the garbage collection timer.
func (t *TokenCache) StopGC() {
	if t.gcTimer != nil {
		t.gcTimer.Stop()
	}
}

// GC removes stale tokens from the cache.
// This function is thread safe and should run periodically.
func (t *TokenCache) GC() {
	t.lock.Lock()
	defer t.lock.Unlock()

	staleTokens := []TokenUID{}
	for id, token := range t.data {
		if time.Now().After(token.expires) {
			staleTokens = append(staleTokens, id)
		}
	}

	// Delete stale tokens in separate loop to avoid invalidating the iterator
	for _, id := range staleTokens {
		delete(t.data, id)
	}
}

// Get reurns the known token for the given service account or nil
// if no token is known or the token has expired.
func (t *TokenCache) Get(tokenIdentifier TokenLookup) *KnownToken {
	t.lock.Lock()
	defer t.lock.Unlock()

	id := tokenIdentifier.ToTokenUID()

	token, isKnown := t.data[id]
	if !isKnown {
		t.missMetric.Inc()
		return nil
	}

	// Remove stale tokens on fetch
	// We assure that a returned token has a minimum lifetime to avoid
	// a token expiring immediately after being fetched.
	if time.Now().Add(t.minTokenLifetime).After(token.expires) {
		log.Debug().TimeDiff("expiresIn", token.expires, time.Now()).Msg("Removing expired, or about to expire, token upon fetch")
		delete(t.data, id)
		t.missMetric.Inc()
		return nil
	}

	if !tokenIdentifier.Equal(token.identifier) {
		log.Warn().Interface("identifier", token.identifier).Msg("Token cache collision detected")
		delete(t.data, id)
		t.missMetric.Inc()
		return nil
	}

	t.hitMetric.Inc()
	return &token
}

// Store stores a token for the given service account.
// The token will be valid until the given time.
func (t *TokenCache) Store(tokenIdentifier TokenLookup, token string, expiresAt time.Time) *KnownToken {
	t.lock.Lock()
	defer t.lock.Unlock()

	storedToken := KnownToken{
		token:      token,
		identifier: tokenIdentifier,
		expires:    expiresAt,
	}

	id := tokenIdentifier.ToTokenUID()
	t.data[id] = storedToken

	t.setMetric.Inc()
	return &storedToken
}

// StoreFor stores a token for the given service account for a given
// duration.
func (t *TokenCache) StoreFor(tokenIdentifier TokenLookup, token string, validFor time.Duration) *KnownToken {
	expireTime := time.Now().Add(validFor)
	return t.Store(tokenIdentifier, token, expireTime)
}

// StoreUntil stores a token for the given service account.
// The expirationDateTime string is expected to be in the format of
// tokenTimeFormat.
func (t *TokenCache) StoreUntil(tokenIdentifier TokenLookup, token, expirationDateTime string) *KnownToken {
	expireTime, err := time.Parse(tokenTimeFormat, expirationDateTime)
	if err != nil {
		return &KnownToken{
			token: token,
		}
	}

	return t.Store(tokenIdentifier, token, expireTime)
}
