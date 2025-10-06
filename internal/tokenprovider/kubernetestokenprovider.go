package tokenprovider

import (
	"context"
	"encoding/json"
	"identity-metadata-server/internal/shared"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
	kubernetes "github.com/trivago/go-kubernetes/v4"
)

const (
	// This value is used for metrics
	kubeAPIendpoint = "kubeapi"
)

// KubernetesTokenProvider is a token provider that uses a kubernetes service account
// to get various different tokens from Google Cloud token endpoints.
type KubernetesTokenProvider struct {
	GcpTokenProvider
	k8s             *kubernetes.Client
	mainAudience    string
	serviceAccounts *KubernetesServiceAccountCache
}

// NewKubernetesTokenProvider creates a new KubernetesToGCPTokenProvider.
func NewKubernetesTokenProvider(workloadIdentityAudience string, client *kubernetes.Client, kubeletHost string, saCacheTTL time.Duration) *KubernetesTokenProvider {
	return &KubernetesTokenProvider{
		GcpTokenProvider: GcpTokenProvider{
			metrics: shared.NewAPIMetrics("metadata_server_k8s", map[string]string{}),
		},
		k8s:             client,
		mainAudience:    workloadIdentityAudience,
		serviceAccounts: NewKubernetesServiceAccountCache(client, kubeletHost, saCacheTTL),
	}
}

// GetIdentityForIP returns information about the service account assigned to
// the pod behind the given IP.
// ServiceAccount data is cached for a short time.
func (tp *KubernetesTokenProvider) GetIdentityForIP(ctx context.Context, ip string) SourceIdentity {
	const metricPath = "identity"

	requestStart := time.Now()
	id := tp.serviceAccounts.Get(ip, ctx)

	statusCode := 200
	if id.owner == nil {
		statusCode = 404
	}

	tp.trackApiCall(kubeAPIendpoint, metricPath, statusCode, requestStart)
	return id
}

// getTokenRequestToken generates a token that can be used to request other
// token types like identity tokens or access tokens.
func (tp *KubernetesTokenProvider) GetTokenRequestToken(ctx context.Context, srcIdentity SourceIdentity, lifetime time.Duration, scopes, additionalAudiences []string) (*shared.TokenExchangeResponse, error) {
	ksa := srcIdentity.(kubernetesServiceAccountInfo)

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
	tokenRequestTokenResponse, err := tp.getSignedRequestToken(lifetime, ksa, scopes, additionalAudiences, ctx)
	if err != nil {
		log.Error().Err(err).
			Str("namespace", ksa.namespace).
			Str("name", ksa.name).
			Str("gsa", ksa.boundGSA).
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
			Str("namespace", ksa.namespace).
			Str("name", ksa.name).
			Str("gsa", ksa.boundGSA).
			Msg("Failed to read token request token")
		return nil, err
	}

	return &tokenRequestToken, nil
}

// getSignedRequestToken returns a signed request token for the given pod IP.
func (tp *KubernetesTokenProvider) getSignedRequestToken(requestTokenLifetime time.Duration, ksa kubernetesServiceAccountInfo, scopes, additionalAudiences []string, ctx context.Context) (*http.Response, error) {
	const metricPath = "request_token"

	// The first audience _has_ to be the workload identity provider.
	audiences := []string{tp.mainAudience}
	if len(additionalAudiences) > 0 {
		audiences = append(audiences, additionalAudiences...)
	}

	requestStart := time.Now()

	// The oidcToken is the token we get from the kubernetes service account.
	// Additional audiences can be added to the subject only, which is the kubernetes sa token in this case.
	// Our main audience is the workload identity provider.
	oidcToken, err := tp.k8s.GetServiceAccountToken(ksa.name, ksa.namespace, requestTokenLifetime, audiences, ksa.owner, ctx)
	tp.trackApiError(kubeAPIendpoint, metricPath, err, requestStart)

	if err != nil {
		log.Error().Str("name", ksa.name).Str("namespace", ksa.namespace).Msg("Failed to get kubernetes service account token")
		return nil, err
	}

	// We need to add the identity token scope if it is not already present.
	// Otherwise we cannot impersonate the service account.
	scopes = shared.AssureIdentityScope(scopes)

	tokenRequest := shared.TokenExchangeRequest{
		Audience:           tp.mainAudience,
		GrantType:          "urn:ietf:params:oauth:grant-type:token-exchange",
		RequestedTokenType: "urn:ietf:params:oauth:token-type:access_token",
		Scope:              strings.Join(scopes, " "), // see endpoint reference below
		SubjectToken:       oidcToken,
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
