package certificates

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

type KeyType int
type KeyStrength int

const (
	// ECDSA represents the Elliptic Curve Digital Signature Algorithm.
	ECDSA KeyType = iota
	// RSA represents the Rivest-Shamir-Adleman algorithm.
	RSA
	// ED25519 represents the Edwards-Curve Digital Signature Algorithm.
	ED25519
)

const (
	// KeyStrengthNormal is recommended for general use.
	KeyStrengthNormal KeyStrength = 1
	// KeyStrengthMedium should be used for enhanced security.
	KeyStrengthMedium KeyStrength = 2
	// KeyStrengthHigh should be used for high-security applications.
	KeyStrengthHigh KeyStrength = 3
)

// CreatePrivateKeyPEM generates a private key in PEM format based on the specified key type.
// It currently supports ECDSA and RSA key types.
// Returns the PEM encoded private key or an error if the key type is unsupported.
func CreatePrivateKeyPEM(t KeyType, s KeyStrength) ([]byte, error) {
	switch t {
	case ED25519:
		// we can use  "golang.org/x/crypto/ed25519" for this, but the standard
		// library does not support x509 encoding for ED25519 keys yet.
		return nil, errors.New("ED25519 key type is not supported yet")
	case ECDSA:
		return CreateECPrivateKeyPEM(s)
	case RSA:
		return CreateRSAPrivateKeyPEM(s)
	default:
		return nil, errors.New("unsupported key type")
	}
}

// CreateECPrivateKeyPEM generates an ECDSA private key in PEM format.
// Returns the PEM encoded private key or an error if the key generation fails.
// The key is generated using a curve appropriate for the requested strength:
// P-256 for normal, P-384 for medium, and P-521 for high strength.
func CreateECPrivateKeyPEM(s KeyStrength) ([]byte, error) {
	var curve elliptic.Curve
	switch s {
	case KeyStrengthNormal:
		curve = elliptic.P256()
	case KeyStrengthMedium:
		curve = elliptic.P384()
	case KeyStrengthHigh:
		curve = elliptic.P521()
	default:
		return nil, errors.New("unsupported key strength")
	}

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}

	privKeyBytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privKeyBytes})
	return privateKeyPEM, nil
}

// CreateRSAPrivateKeyPEM generates an RSA private key in PEM format.
// Returns the PEM encoded private key or an error if the key generation fails.
// The key is generated with a bit length based on the requested strength:
// 2048 for normal, 3072 for medium, and 4096 for high strength.
func CreateRSAPrivateKeyPEM(s KeyStrength) ([]byte, error) {
	var bits int
	switch s {
	case KeyStrengthNormal:
		bits = 2048
	case KeyStrengthMedium:
		bits = 3072
	case KeyStrengthHigh:
		bits = 4096
	default:
		return nil, errors.New("unsupported key strength")
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: privKeyBytes})
	return privateKeyPEM, nil
}
