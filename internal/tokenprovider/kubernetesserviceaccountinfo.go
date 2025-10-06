package tokenprovider

import (
	"hash"
	"strings"
	"time"

	"github.com/cespare/xxhash"
	"github.com/trivago/go-kubernetes/v4"
)

// kubernetesServiceAccountInfo holds information about a kubernetes service account.
// It implements the ServiceIdentity interface.
type kubernetesServiceAccountInfo struct {
	name      string
	namespace string
	boundGSA  string
	owner     kubernetes.NamedObject
	firstSeen time.Time
}

// Hash returns a hash of the service account information.
func (ksa kubernetesServiceAccountInfo) Hash() hash.Hash64 {
	idString := strings.Join([]string{ksa.namespace, ksa.name, ksa.boundGSA}, ";")
	idHash := xxhash.New()
	idHash.Write([]byte(idString))
	return idHash
}

// GetBoundGSA returns the bound GSA for the service account.
func (ksa kubernetesServiceAccountInfo) GetBoundGSA() string {
	return ksa.boundGSA
}

// Equal compares two host identities.
// We don't compare the owner, as it is not relevant for the identity
// (see Hash function).
func (ksa kubernetesServiceAccountInfo) Equal(other SourceIdentity) bool {
	ksa2, isSameType := other.(kubernetesServiceAccountInfo)
	return isSameType &&
		ksa.name == ksa2.name &&
		ksa.namespace == ksa2.namespace &&
		ksa.boundGSA == ksa2.boundGSA
}
