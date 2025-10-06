package shared

// https://cloud.google.com/iam/docs/reference/sts/rest/v1/TopLevel/token#request-body
type TokenExchangeRequest struct {
	GrantType          string `json:"grantType,omitempty"`
	Audience           string `json:"audience,omitempty"`
	Scope              string `json:"scope,omitempty"`
	RequestedTokenType string `json:"requestedTokenType,omitempty"`
	SubjectToken       string `json:"subjectToken,omitempty"`
	SubjectTokenType   string `json:"subjectTokenType,omitempty"`
	Options            string `json:"options,omitempty"`
	LifetimeSec        string `json:"lifetime,omitempty"`
}

// As defined in the identity server
type HostTokenRequest struct {
	Audiences []string `json:"audiences"`
	Lifetime  string   `json:"lifetime,omitempty"`
}

// https://cloud.google.com/iam/docs/reference/sts/rest/v1/TopLevel/token#response-body
type TokenExchangeResponse struct {
	AccessToken              string `json:"access_token,omitempty"`
	ExpiresIn                int    `json:"expires_in,omitempty"`
	TokenType                string `json:"token_type,omitempty"`
	IssuedTokenType          string `json:"issued_token_type,omitempty"`
	AccessBoundarySessionKey string `json:"access_boundary_session_key,omitempty"`
}

// https://cloud.google.com/iam/docs/reference/credentials/rest/v1/projects.serviceAccounts/generateAccessToken#request-body
type IAMAccessTokenRequest struct {
	Scope       []string `json:"scope,omitempty"`
	LifetimeSec string   `json:"lifetime,omitempty"`
}

// https://cloud.google.com/iam/docs/reference/credentials/rest/v1/projects.serviceAccounts/generateAccessToken#response-body
type IAMAccessTokenResponse struct {
	AccessToken string `json:"accessToken,omitempty"`
	ExpireTime  string `json:"expireTime,omitempty"`
}

// https://cloud.google.com/iam/docs/reference/credentials/rest/v1/projects.serviceAccounts/generateIdToken#request-body
type IAMIdentityTokenRequest struct {
	Delegates    []string `json:"delegates,omitempty"`
	Audience     string   `json:"audience,omitempty"`
	IncludeEmail bool     `json:"includeEmail,omitempty"`
}

// https://cloud.google.com/iam/docs/reference/credentials/rest/v1/projects.serviceAccounts/generateIdToken#response-body
type IAMIdentityTokenResponse struct {
	Token string `json:"token,omitempty"`
}
