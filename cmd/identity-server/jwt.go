package main

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// NodeClaims holds fields we use to identify a node.
// These claims have to be unique for each node.
// For Workload Identity Federation, these fields need to be
// the same as the ones used in the workload identity pool.
type NodeClaims struct {
	Identity string `json:"identity"`
}

// CustomClaims holds the claims we want to include in the JWT.
type CustomClaims struct {
	NodeClaims NodeClaims `json:"node_claims"`
	jwt.RegisteredClaims
}

type signingKey struct {
	signingKey interface{}
	setKey     jwk.Key
	method     jwt.SigningMethod
}

var (
	jwks      jwk.Set
	serverKey *signingKey
)

func newSigningKey(setKey jwk.Key, nativeKey interface{}, method jwt.SigningMethod) *signingKey {
	sk := signingKey{
		signingKey: nativeKey,
		setKey:     setKey,
		method:     method,
	}

	keyId := viper.GetString("server.keyName")

	// Set required fields
	_ = sk.setKey.Set(jwk.KeyUsageKey, "sig")
	_ = sk.setKey.Set(jwk.KeyIDKey, keyId)

	// Remove unsupported fields
	_ = sk.setKey.Remove("d")
	_ = sk.setKey.Remove("dp")
	_ = sk.setKey.Remove("dq")
	_ = sk.setKey.Remove("p")
	_ = sk.setKey.Remove("q")
	_ = sk.setKey.Remove("qi")

	return &sk
}

func readJwkKeyFromPKCS8(data []byte) (*signingKey, bool, error) {
	key, err := x509.ParsePKCS8PrivateKey(data)
	if err != nil {
		return nil, false, err
	}

	switch nativeKey := key.(type) {
	case *rsa.PrivateKey:
		setKey := jwk.NewRSAPrivateKey()
		if err := setKey.FromRaw(nativeKey); err != nil {
			return nil, true, err
		}

		return newSigningKey(setKey, nativeKey, jwt.SigningMethodRS256), true, nil

	case *ecdsa.PrivateKey:
		setKey := jwk.NewECDSAPrivateKey()
		if err := setKey.FromRaw(nativeKey); err != nil {
			return nil, true, err
		}
		return newSigningKey(setKey, nativeKey, jwt.SigningMethodES256), true, nil

	default:
		return nil, true, fmt.Errorf("unsupported key type")
	}
}

func readJwkKeyFromPKCS1(data []byte) (*signingKey, bool, error) {
	rsaKey, err := x509.ParsePKCS1PrivateKey(data)
	if err != nil {
		return nil, false, err
	}

	setKey := jwk.NewRSAPrivateKey()
	if err := setKey.FromRaw(rsaKey); err != nil {
		return nil, true, err
	}

	return newSigningKey(setKey, rsaKey, jwt.SigningMethodRS256), true, nil
}

func readJwkKeyFromEC(data []byte) (*signingKey, bool, error) {
	ecKey, err := x509.ParseECPrivateKey(data)
	if err != nil {
		return nil, false, err
	}

	setKey := jwk.NewECDSAPrivateKey()
	if err := setKey.FromRaw(ecKey); err != nil {
		return nil, true, err
	}

	return newSigningKey(setKey, ecKey, jwt.SigningMethodES256), true, nil
}

// initJWKS generates a JWKS from the server's private key.
// Calling this function will overwrite the global jwks variable, hence it can
// also be used to reload the JWKS.
func initJWKS() error {
	// Read the private key from disk
	keyPath := viper.GetString("server.key")
	rawKey, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}

	block, _ := pem.Decode(rawKey)
	if block == nil || len(block.Bytes) == 0 {
		return fmt.Errorf("failed to decode PEM data from file %s", keyPath)
	}

	parser := []func([]byte) (*signingKey, bool, error){
		readJwkKeyFromPKCS8,
		readJwkKeyFromPKCS1,
		readJwkKeyFromEC,
	}

	for _, p := range parser {
		candidateKey, formatMatched, err := p(block.Bytes)
		if !formatMatched {
			continue
		}
		if err != nil {
			log.Error().Err(err).Msg("Failed to parse private key")
			return err
		}

		// The key is in a supported format
		serverKey = candidateKey
		jwks = jwk.NewSet()
		jwks.Add(serverKey.setKey)
		return nil
	}

	return fmt.Errorf("failed to parse private key: unsupported format")
}

func buildAndSignJWT(claims CustomClaims) (string, error) {
	if serverKey == nil {
		return "", ErrorSigningKeyNotLoaded
	}

	keyId := viper.GetString("server.keyName")

	token := jwt.NewWithClaims(serverKey.method, claims)
	token.Header["kid"] = keyId
	return token.SignedString(serverKey.signingKey)
}
