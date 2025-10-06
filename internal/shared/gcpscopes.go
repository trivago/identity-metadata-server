package shared

import "slices"

const (
	DefaultScope       = "https://www.googleapis.com/auth/cloud-platform"
	IdentityTokenScope = "https://www.googleapis.com/auth/iam"

	EndpointIAMCredentials = "iamcredentials.googleapis.com/v1"
	EndpointSTS            = "sts.googleapis.com/v1"
)

// AssureIdentityScope checks if the IdentityTokenScope is present in the provided scopes.
// If not, it adds it to the beginning of the scopes slice.
// This is necessary for the token exchange to work properly when using Workload Identity.
func AssureIdentityScope(scopes []string) []string {
	if !slices.Contains(scopes, DefaultScope) && !slices.Contains(scopes, IdentityTokenScope) {
		return append([]string{IdentityTokenScope}, scopes...)
	}
	return scopes
}

// GetWorkloadIdentityAudience returns the workload identity pool audience for
// a given project number, pool name, and provider name.
func GetWorkloadIdentityAudience(projectNumber, poolName, providerName string) string {
	return "//iam.googleapis.com/projects/" + projectNumber + "/locations/global/workloadIdentityPools/" + poolName + "/providers/" + providerName
}
