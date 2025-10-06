package main

import (
	"fmt"
	"identity-metadata-server/internal/tokenprovider"
	"strings"

	"k8s.io/utils/strings/slices"
)

type TokenLookup struct {
	Type                TokenType
	Identity            tokenprovider.SourceIdentity
	Scopes              []string
	AdditionalAudiences []string
}

// TokenUID is a unique identifier for a service account
type TokenUID string

// TokenType denotes the type of token
type TokenType string

const (
	TokenTypeAccess   TokenType = "access"
	TokenTypeIdentity TokenType = "id"
)

// NewLookup creates a new TokenLookup from a kubernetesServiceAccountInfo
// without scope or audience
func NewLookup(tokenType TokenType, srcIdentity tokenprovider.SourceIdentity) TokenLookup {
	return TokenLookup{
		Type:     tokenType,
		Identity: srcIdentity,
	}
}

// NewLookupWithScopeAndAudience creates a new TokenLookup from a kubernetesServiceAccountInfo
// with a scope and an audience. This is typically used for access tokens.
func NewLookupWithScopeAndAudience(tokenType TokenType, srcIdentity tokenprovider.SourceIdentity, scopes, additionalAudiences []string) TokenLookup {
	return TokenLookup{
		Type:                tokenType,
		Identity:            srcIdentity,
		Scopes:              scopes,
		AdditionalAudiences: additionalAudiences,
	}
}

// NewLookupWithAudience creates a new TokenLookup from a kubernetesServiceAccountInfo
// with a single audience and no scope. This is typically used for identity tokens.
func NewLookupWithAudience(tokenType TokenType, srcIdentity tokenprovider.SourceIdentity, audience string) TokenLookup {
	return TokenLookup{
		Type:                tokenType,
		Identity:            srcIdentity,
		AdditionalAudiences: []string{audience},
	}
}

// ToTokenUID converts a TokenLookup to a serviceAccountID that can be
// used to retreive a token from the cache
func (t TokenLookup) ToTokenUID() TokenUID {
	idHash := t.Identity.Hash()
	hashExt := strings.Join(append(t.AdditionalAudiences, t.Scopes...), ";")
	idHash.Write([]byte(hashExt))

	return TokenUID(fmt.Sprintf("%s:%d", t.Type, idHash.Sum64()))
}

// Equal compares two TokenLookups for equality
func (t TokenLookup) Equal(t2 TokenLookup) bool {
	return t.Type == t2.Type &&
		t.Identity.Equal(t2.Identity) &&
		EqualUnordered(t.Scopes, t2.Scopes) &&
		EqualUnordered(t.AdditionalAudiences, t2.AdditionalAudiences)
}

// EqualUnordered compares two string slices for equality
// without considering the order of elements.
func EqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, v := range a {
		if !slices.Contains(b, v) {
			return false
		}
	}
	return true
}
