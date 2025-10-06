package certificates

import (
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePrivateKeyPEM(t *testing.T) {
	assert := assert.New(t)

	key, err := CreatePrivateKeyPEM(ECDSA, KeyStrengthNormal)
	assert.NoError(err)
	assert.NotNil(key)

	keyBlock, _ := pem.Decode(key)
	assert.Equal("EC PRIVATE KEY", keyBlock.Type)

	// Test RSA key generation
	rsaKey, err := CreatePrivateKeyPEM(RSA, KeyStrengthNormal)
	assert.NoError(err)
	assert.NotNil(rsaKey)

	rsaKeyBlock, _ := pem.Decode(rsaKey)
	assert.Equal("RSA PRIVATE KEY", rsaKeyBlock.Type)
}
