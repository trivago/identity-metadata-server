package tokenprovider

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"identity-metadata-server/internal/certificates"
	"identity-metadata-server/internal/shared"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cespare/xxhash"
	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

// KubernetesTokenProvider is a token provider that uses a kubernetes service account
// to get various different tokens from Google Cloud token endpoints.
type HostTokenProvider struct {
	GcpTokenProvider
	mainAudience    string
	serverUrl       string
	cachedIdentity  hostIdentity
	clientCertPath  string
	clientKeyPath   string
	certificate     tls.Certificate
	certMinLifetime time.Duration
	refreshCertTick *time.Ticker
	tickerDone      chan struct{}

	identityGuard *sync.Mutex
}

type hostIdentity struct {
	BoundGSA string
}

// Equal compares two host identities.
func (h hostIdentity) Equal(other SourceIdentity) bool {
	h2, isSameType := other.(hostIdentity)
	return isSameType && h.BoundGSA == h2.BoundGSA
}

// NewKubernetesTokenProvider creates a new KubernetesToGCPTokenProvider.
func NewHostTokenProvider(workloadIdentityAudience, identityServerURL, caCertPath, clientCertPath, clientKeyPath string, refreshInterval, clientCertMinLifetime time.Duration) (*HostTokenProvider, error) {
	if caCertPath != "" {
		// Load the CA certificate
		// The certificate is expected to be in PEM format
		if err := shared.RegisterRootCAFile(caCertPath); err != nil {
			log.Error().Err(err).Msg("Failed to read CA certificate")
		}
	}

	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed to load client certificate"))
	}

	// When compiling with go before 1.23, clientCert.Leaf is always nil.
	// After go 1.23 this will also happen if you set "x509keypairleaf=0" in the GODEBUG
	// environment variable. To assure we always have a valid leaf certificate, we need to
	// parse the certificate manually if Leaf is nil.
	if clientCert.Leaf == nil {
		clientCert.Leaf, err = x509.ParseCertificate(clientCert.Certificate[0])
		if err != nil {
			return nil, errors.Join(err, errors.New("failed to parse client certificate"))
		}
	}

	certLifetime := clientCert.Leaf.NotAfter.Sub(clientCert.Leaf.NotBefore)
	if certLifetime <= clientCertMinLifetime {
		return nil, fmt.Errorf("client certificate min lifetime of %s is shorter than or equal to the total certificate lifetime of %s", clientCertMinLifetime.String(), certLifetime.String())
	}

	provider := &HostTokenProvider{
		GcpTokenProvider: GcpTokenProvider{
			metrics: shared.NewAPIMetrics("metadata_server_host", map[string]string{}),
		},
		mainAudience:    workloadIdentityAudience,
		serverUrl:       identityServerURL,
		certificate:     clientCert,
		clientCertPath:  clientCertPath,
		clientKeyPath:   clientKeyPath,
		identityGuard:   new(sync.Mutex),
		refreshCertTick: time.NewTicker(refreshInterval),
		tickerDone:      make(chan struct{}),
		certMinLifetime: clientCertMinLifetime,
	}

	// Always try to refresh the certificate on startup
	err = provider.TryRefreshCertificate()
	if err != nil {
		// If the first refresh fails, we stop the ticker, as no provider is returned
		provider.refreshCertTick.Stop()
		return nil, err
	}

	// Make sure we check on the certificate every 24 hours
	go func() {
		for {
			select {
			case <-provider.tickerDone:
				provider.refreshCertTick.Stop()
				return
			case <-provider.refreshCertTick.C:
				if err := provider.TryRefreshCertificate(); err != nil {
					log.Error().Err(err).Msg("Failed to refresh certificate")
				}
			}
		}
	}()

	return provider, nil
}

// Close stops the certificate refresh ticker.
// It is important to call this method to avoid memory leaks.
func (tp *HostTokenProvider) Close() {
	// We don't call Stop() on the ticker, but instead close the guard channel
	// to let the goroutine stop the ticker. Stop() does not close the ticker
	// channel, i.e. the goroutine would block forever.
	close(tp.tickerDone)
}

// ClearIdentityCache clears the cached identity.
func (tp *HostTokenProvider) ClearIdentityCache() {
	tp.identityGuard.Lock()
	defer tp.identityGuard.Unlock()
	tp.cachedIdentity = hostIdentity{}
}

// GetIdentityForIP returns information about the service account assigned to
// the current Host. The given ip is ignored.
// The identity is returned from cache after the first successful request.
// If the identity could not be retrieved, an empty identity is returned.
func (tp *HostTokenProvider) GetIdentityForIP(ctx context.Context, ip string) SourceIdentity {
	tp.identityGuard.Lock()
	defer tp.identityGuard.Unlock()

	const metricPath = "identity"

	if len(tp.cachedIdentity.BoundGSA) > 0 {
		return tp.cachedIdentity
	}

	requestStart := time.Now()
	rsp, err := shared.HttpGET(tp.serverUrl+"/identity", nil, nil, &tp.certificate, ctx)

	tp.trackApiResponse(tp.serverUrl, metricPath, rsp, requestStart)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get identity for current host")
		return hostIdentity{}
	}

	defer func() { _ = rsp.Body.Close() }()
	body, _ := io.ReadAll(rsp.Body)

	if rsp.StatusCode != http.StatusOK {
		log.Error().
			Int("status", rsp.StatusCode).
			Str("message", string(body)).
			Msg("Failed to get identity for current host")
		return hostIdentity{}
	}

	returnedIdentities := strings.Split(string(body), "\n")
	if len(returnedIdentities) == 0 {
		log.Error().Str("body", string(body)).Msg("No identity returned for current host")
		return hostIdentity{}
	}

	boundIdentity := strings.TrimSpace(returnedIdentities[0])
	if len(boundIdentity) == 0 {
		log.Error().Str("body", string(body)).Msg("Empty identity returned for current host")
		return hostIdentity{}
	}

	tp.cachedIdentity = hostIdentity{
		BoundGSA: boundIdentity,
	}
	return tp.cachedIdentity
}

// getTokenRequestToken generates a token that can be used to request other
// token types like identity tokens or access tokens.
func (tp *HostTokenProvider) GetTokenRequestToken(ctx context.Context, srcIdentity SourceIdentity, lifetime time.Duration, scopes, additionalAudiences []string) (*shared.TokenExchangeResponse, error) {
	machineIdentity := srcIdentity.(hostIdentity)
	if len(machineIdentity.BoundGSA) == 0 {
		return nil, shared.WrapErrorWithStatus(fmt.Errorf("failed to get bound GSA for current host"), http.StatusUnauthorized)
	}

	// Warning: requestTokenLifetime must not be less than 10 minutes
	// The corresponding error message is:
	// Invalid value: 10: may not specify a duration less than 10 minutes
	if lifetime < time.Minute*10 {
		log.Warn().Dur("lifetime", lifetime).Msg("Request token lifetime is clamped to 10 minutes minimum")
		lifetime = time.Minute * 10
	}

	tokenRequestToken := shared.TokenExchangeResponse{}

	// Basically the same steps as in getAccessToken, but
	// we need to access the different fields
	tokenRequestTokenResponse, err := tp.getSignedRequestToken(ctx, lifetime, scopes, additionalAudiences)
	if err != nil {
		log.Error().Err(err).
			Str("gsa", machineIdentity.BoundGSA).
			Msg("Failed to get access token")
		return nil, err
	}
	defer func() { _ = tokenRequestTokenResponse.Body.Close() }()

	// If the response is not 200, we log the response and return nil
	if tokenRequestTokenResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(tokenRequestTokenResponse.Body)
		log.Error().
			Int("status", tokenRequestTokenResponse.StatusCode).
			Str("content-type", tokenRequestTokenResponse.Header.Get("Content-Type")).
			Str("body", string(body)).
			Msg("Failed to get access token")
		return nil, shared.WrapErrorWithStatus(err, tokenRequestTokenResponse.StatusCode)
	}

	if err := json.NewDecoder(tokenRequestTokenResponse.Body).Decode(&tokenRequestToken); err != nil {
		log.Error().Err(err).
			Str("gsa", machineIdentity.BoundGSA).
			Msg("Failed to read token request token")
		return nil, err
	}

	return &tokenRequestToken, nil
}

// getSignedRequestToken returns a signed request token for the given pod IP.
func (tp *HostTokenProvider) getSignedRequestToken(ctx context.Context, requestTokenLifetime time.Duration, scopes, additionalAudiences []string) (*http.Response, error) {
	const metricPath = "request_token"

	// The first audience _has_ to be the workload identity provider.
	audiences := []string{tp.mainAudience}
	if len(additionalAudiences) > 0 {
		audiences = append(audiences, additionalAudiences...)
	}

	identityTokenRequest, err := jsoniter.Marshal(shared.HostTokenRequest{
		Audiences: audiences,
		Lifetime:  requestTokenLifetime.String(),
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal identity server token request")
		return nil, shared.WrapErrorWithStatus(err, http.StatusBadRequest)
	}

	requestStart := time.Now()
	oidcTokenRsp, err := shared.HttpGET(tp.serverUrl+"/token", identityTokenRequest, nil, &tp.certificate, ctx)

	tp.trackApiResponse(tp.serverUrl, metricPath, oidcTokenRsp, requestStart)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get identity token")
		return nil, shared.WrapErrorWithStatus(err, http.StatusInternalServerError)
	}

	defer func() { _ = oidcTokenRsp.Body.Close() }()
	oidcTokenBody, err := io.ReadAll(oidcTokenRsp.Body)

	if err != nil {
		log.Error().Err(err).Msg("Failed to read identity token body")
		return nil, shared.WrapErrorWithStatus(err, http.StatusInternalServerError)
	}

	if oidcTokenRsp.StatusCode != http.StatusOK {
		err := fmt.Errorf("%s", string(oidcTokenBody))
		log.Error().Err(err).Msg("Identity token request failed")
		return nil, shared.WrapErrorWithStatus(err, oidcTokenRsp.StatusCode)
	}

	// We need to add the identity token scope if it is not already present.
	// Otherwise we cannot impersonate the service account.
	scopes = shared.AssureIdentityScope(scopes)

	tokenRequest := shared.TokenExchangeRequest{
		Audience:           tp.mainAudience,
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
		Scope:              strings.Join(scopes, " "), // see endpoint reference below
		SubjectToken:       string(oidcTokenBody),
		SubjectTokenType:   "urn:ietf:params:oauth:token-type:jwt",
		LifetimeSec:        strconv.Itoa(int(requestTokenLifetime.Seconds())),
	}

	tokenRequestBody, err := jsoniter.MarshalToString(tokenRequest)
	if err != nil {
		log.Error().Msg("Failed to marshal token request")
		return nil, shared.WrapErrorWithStatus(err, http.StatusBadRequest)
	}

	requestStart = time.Now()

	// See https://cloud.google.com/iam/docs/reference/sts/rest/v1/TopLevel/token
	rsp, err := http.Post("https://"+shared.EndpointSTS+"/token",
		"application/json",
		strings.NewReader(tokenRequestBody))

	tp.trackApiResponse(shared.EndpointSTS, metricPath, rsp, requestStart)
	return rsp, err
}

func (tp *HostTokenProvider) TryRefreshCertificate() error {
	const metricPath = "renew"

	// As we're calling this function in long intervals, a short lock like this
	// is currently considered ok.
	// If we ever expose this function via a REST API, we need to
	// make sure we have a proper locking mechanism in place to avoid
	// this function being called multiple times in parallel.
	tp.identityGuard.Lock()
	oldCert := tp.certificate.Leaf
	tp.identityGuard.Unlock()

	remaining := time.Until(oldCert.NotAfter)
	if remaining <= 0 {
		return fmt.Errorf("client certificate already expired %s ago. A manual refresh is needed", -remaining)
	}

	if remaining > tp.certMinLifetime {
		log.Info().Msg("Certificate is still valid, no need to refresh")
		return nil
	}

	// Either get a new private key from disk or create a new one
	fileSuffix := oldCert.NotAfter.Add(-tp.certMinLifetime).Format("20060102150405")
	keyBasePath := filepath.Dir(tp.clientKeyPath)
	keyFilePath := filepath.Join(keyBasePath, fmt.Sprintf("client.key.%s", fileSuffix))

	var privateKeyPEM []byte

	_, err := os.Stat(keyFilePath)
	switch {
	case err == nil:
		privateKeyPEM, err = os.ReadFile(keyFilePath)
		if err != nil {
			return errors.Join(err, errors.New("failed to read private key"))
		}

	case os.IsNotExist(err):
		// Create a new private key
		privateKeyPEM, err = certificates.CreatePrivateKeyPEM(certificates.ECDSA, certificates.KeyStrengthMedium)
		if err != nil {
			return errors.Join(err, errors.New("failed to create private key"))
		}
		log.Info().Str("path", keyFilePath).Msg("Writing new private key to disk")

		// Write the private key to disk so we can use it laster
		if err := os.WriteFile(keyFilePath, privateKeyPEM, 0600); err != nil {
			return errors.Join(err, errors.New("failed to write private key"))
		}

	default:
		return errors.Join(err, errors.New("failed to check private key"))
	}

	// Create a CSR from the current certificate
	csr, err := certificates.CreateClientCSRFromCertificate(privateKeyPEM, tp.certificate.Leaf)
	if err != nil {
		return errors.Join(err, errors.New("failed to create CSR"))
	}

	// Get a new certificate from the identity server
	requestStart := time.Now()
	rsp, err := shared.HttpPOST(tp.serverUrl+"/renew", []byte(csr), map[string]string{
		"Content-Type": "application/x-pem-file",
		"Accept":       "application/x-pem-file",
	}, &tp.certificate, context.Background())

	tp.trackApiResponse(tp.serverUrl, metricPath, rsp, requestStart)
	if err != nil {
		return errors.Join(err, errors.New("failed to get new certificate"))
	}

	defer func() { _ = rsp.Body.Close() }()
	if rsp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(rsp.Body)
		return errors.Join(errors.New(string(body)), errors.New("failed to get new certificate from identity server"))
	}

	newCertPEM, err := io.ReadAll(rsp.Body)
	if err != nil {
		return errors.Join(err, errors.New("failed to read new certificate"))
	}

	// Create a new keypair. This will also validate the certificate
	// and make sure it is valid.
	clientCert, err := tls.X509KeyPair(newCertPEM, privateKeyPEM)
	if err != nil {
		return errors.Join(err, errors.New("failed to create new client certificate"))
	}

	// Write the new certificate to disk
	certBasePath := filepath.Dir(tp.clientCertPath)
	certFilePath := filepath.Join(certBasePath, fmt.Sprintf("client.cert.%s", fileSuffix))

	log.Info().Str("path", certFilePath).Msg("Writing new client certificate to disk")

	if err := os.WriteFile(certFilePath, newCertPEM, 0644); err != nil {
		return errors.Join(err, errors.New("failed to write new client certificate"))
	}

	// Rotate the symlinks for the client certificate and key.
	// If any of the rotation fails, the changes will be rolled back.

	rotateFiles := shared.NewKVList[string, string]()
	rotateFiles.Add(tp.clientCertPath, certFilePath)
	rotateFiles.Add(tp.clientKeyPath, keyFilePath)

	if err := shared.RotateSymlinkList(rotateFiles); err != nil {
		return errors.Join(err, errors.New("failed to rotate symlinks for new certificate"))
	}

	// Everything is fine, we can use the new certificate.
	// Make sure to use the identity guard so we wait for
	// any ongoing identity server requests to finish.
	tp.identityGuard.Lock()
	defer tp.identityGuard.Unlock()

	tp.certificate = clientCert
	return nil
}

// Hash returns a hash of the service account information.
func (h hostIdentity) Hash() hash.Hash64 {
	idHash := xxhash.New()
	idHash.Write([]byte(h.BoundGSA))
	return idHash
}

// GetBoundGSA returns the bound GSA for the service account.
func (h hostIdentity) GetBoundGSA() string {
	return h.BoundGSA
}
