package tokenprovider

import (
	"context"
	"fmt"
	"identity-metadata-server/internal/shared"
	"io"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/rs/zerolog/log"
)

// GCPTokenProvider implements those parts of the TokenProvider interface that
// are specific to GCP and don't require knowledge of the original token.
type GcpTokenProvider struct {
	metrics *shared.APIMetrics
}

// GetAccessToken tries to get an access token for the given scope and GSA.
func (tp *GcpTokenProvider) GetAccessToken(ctx context.Context, tokenRequestToken shared.TokenExchangeResponse, lifetime time.Duration, scopes []string, gsa string) (*shared.IAMAccessTokenResponse, error) {
	const metricPath = "access_token"

	// We request the proper GCP auth token using the tokenRequestToken we got
	// https://cloud.google.com/iam/docs/create-short-lived-credentials-direct#create-access
	accessTokenRequest := shared.IAMAccessTokenRequest{
		Scope:       scopes,
		LifetimeSec: fmt.Sprintf("%ds", int(lifetime.Seconds())),
	}

	accessTokenRequestBody, err := jsoniter.Marshal(accessTokenRequest)
	if err != nil {
		log.Error().Err(err).
			Str("scopes", strings.Join(scopes, ",")).
			Str("gsa", gsa).
			Msg("Failed to marshal gcp token request")
		return nil, shared.WrapErrorWithStatus(err, http.StatusBadRequest)
	}

	// See https://cloud.google.com/iam/docs/reference/credentials/rest/v1/projects.serviceAccounts/generateAccessToken
	requestStart := time.Now()
	accessTokenResponse, err := shared.HttpPOST(
		"https://"+shared.EndpointIAMCredentials+"/projects/-/serviceAccounts/"+gsa+":generateAccessToken",
		accessTokenRequestBody,
		map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + tokenRequestToken.AccessToken,
		}, nil, ctx)

	tp.trackApiResponse(shared.EndpointIAMCredentials, metricPath, accessTokenResponse, requestStart)
	if err != nil {
		log.Error().Err(err).
			Str("scopes", strings.Join(scopes, ",")).
			Str("gsa", gsa).
			Msg("Failed to call iam access token endpoint")
		return nil, shared.WrapErrorWithStatus(err, http.StatusInternalServerError)
	}
	defer accessTokenResponse.Body.Close()

	// If the response is not 200, we log the response and return nil
	if accessTokenResponse.StatusCode != http.StatusOK {
		// make sure to give the KSA Workload Identity User permissions on the target GSA
		// principal://iam.googleapis.com/projects/{{project number}}/locations/global/workloadIdentityPools/{{pool}}/subject/system:serviceaccount:{{namespace}}:{{ksa}}

		body, _ := io.ReadAll(accessTokenResponse.Body)
		log.Error().
			Str("scopes", strings.Join(scopes, ",")).
			Str("gsa", gsa).
			Int("status", accessTokenResponse.StatusCode).
			Str("content-type", accessTokenResponse.Header.Get("Content-Type")).
			Str("body", string(body)).
			Msg("credentials endpoint returned a non-200 status")

		return nil, shared.WrapErrorWithStatus(err, accessTokenResponse.StatusCode)
	}

	gcpAccessToken := shared.IAMAccessTokenResponse{}
	err = jsoniter.NewDecoder(accessTokenResponse.Body).Decode(&gcpAccessToken)
	if err != nil {
		log.Error().Err(err).
			Str("scopes", strings.Join(scopes, ",")).
			Str("gsa", gsa).
			Msg("Failed to get decode gcp token")
		return nil, shared.WrapErrorWithStatus(err, http.StatusInternalServerError)
	}

	return &gcpAccessToken, nil
}

// getIdentityToken tries to get an identity token for the given audience and GSA.
func (tp *GcpTokenProvider) GetIdentityToken(ctx context.Context, tokenRequestToken shared.TokenExchangeResponse, gsa string, audience string) (*shared.IAMIdentityTokenResponse, error) {
	const metricPath = "id_token"

	// We request the proper GCP auth token using the tokenRequestToken we got
	// https://cloud.google.com/iam/docs/create-short-lived-credentials-direct#create-access
	identityTokenRequest := shared.IAMIdentityTokenRequest{
		Audience:     audience,
		IncludeEmail: true,
	}

	identityTokenRequestBody, err := jsoniter.Marshal(identityTokenRequest)
	if err != nil {
		log.Error().Err(err).
			Str("gsa", gsa).
			Str("audience", audience).
			Msg("Failed to marshal identity token request")
		return nil, shared.WrapErrorWithStatus(err, http.StatusBadRequest)
	}

	// See https://cloud.google.com/iam/docs/reference/credentials/rest/v1/projects.serviceAccounts/generateIdToken
	requestStart := time.Now()
	identityTokenResponse, err := shared.HttpPOST(
		"https://"+shared.EndpointIAMCredentials+"/projects/-/serviceAccounts/"+gsa+":generateIdToken",
		identityTokenRequestBody,
		map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer " + tokenRequestToken.AccessToken,
		}, nil, ctx)

	tp.trackApiResponse(shared.EndpointIAMCredentials, metricPath, identityTokenResponse, requestStart)
	if err != nil {
		log.Error().Err(err).
			Msg("Failed to call iam identity token endpoint")
		return nil, shared.WrapErrorWithStatus(err, http.StatusInternalServerError)
	}
	defer identityTokenResponse.Body.Close()

	// If the response is not 200, we log the response and return nil
	if identityTokenResponse.StatusCode != http.StatusOK {
		// make sure to give the KSA Workload Identity User permissions on the target GSA
		// principal://iam.googleapis.com/projects/{{project number}}/locations/global/workloadIdentityPools/{{pool}}/subject/system:serviceaccount:{{namespace}}:{{ksa}}

		body, _ := io.ReadAll(identityTokenResponse.Body)
		log.Error().
			Str("gsa", gsa).
			Str("audience", audience).
			Int("status", identityTokenResponse.StatusCode).
			Str("content-type", identityTokenResponse.Header.Get("Content-Type")).
			Str("body", string(body)).
			Msg("credentials endpoint returned a non-200 status")

		return nil, shared.WrapErrorWithStatus(err, identityTokenResponse.StatusCode)
	}

	identityToken := shared.IAMIdentityTokenResponse{}
	err = jsoniter.NewDecoder(identityTokenResponse.Body).Decode(&identityToken)
	if err != nil {
		log.Error().Err(err).
			Str("gsa", gsa).
			Str("audience", audience).
			Msg("Failed to get decode identity token")
		return nil, shared.WrapErrorWithStatus(err, http.StatusInternalServerError)
	}

	return &identityToken, nil
}

// trackApiResponse is a wrapper around trackApiCall that extracts the status code
// from the http.Response unless it is nil.
func (tp *GcpTokenProvider) trackApiResponse(endpoint, path string, rsp *http.Response, requestStart time.Time) {
	statusCode := -1
	if rsp != nil {
		statusCode = rsp.StatusCode
	}

	tp.trackApiCall(endpoint, path, statusCode, requestStart)
}

// trackApiError is a wrapper around trackApiCall that extracts the status code
// from the error if possible. No error will yield status code 200, an unknown error
// will yield status code -1, and an error with a known status code will yield that code.
func (tp *GcpTokenProvider) trackApiError(endpoint, path string, err error, requestStart time.Time) {
	if err == nil {
		tp.trackApiCall(endpoint, path, http.StatusOK, requestStart)
	} else if httpErr, ok := err.(*shared.ErrorWithStatus); ok {
		tp.trackApiCall(endpoint, path, httpErr.Code, requestStart)
	} else {
		tp.trackApiCall(endpoint, path, -1, requestStart)
	}
}

// trackApiCall is a helper function to track the request duration and status code
// for the given endpoint and path. If metrics are not initialized, it does nothing.
func (tp *GcpTokenProvider) trackApiCall(endpoint, path string, statusCode int, requestStart time.Time) {
	if tp.metrics == nil {
		return
	}

	tp.metrics.TrackDuration(endpoint, path, time.Since(requestStart))
	tp.metrics.TrackRequest(endpoint, path, statusCode)
}
