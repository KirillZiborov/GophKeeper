package encryption_test

import (
	"testing"

	"github.com/KirillZiborov/GophKeeper/pkg/encryption"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptDecryptWithKey(t *testing.T) {
	key := "my-secret-key"
	plaintext := "This is a secret message"

	encrypted, err := encryption.EncryptWithKey(plaintext, key)
	require.NoError(t, err)
	require.NotEmpty(t, encrypted, "Encrypted data should not be empty")

	decrypted, err := encryption.DecryptWithKey(encrypted, key)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted, "Decrypted text should match the original")
}

func TestHashUnhash(t *testing.T) {
	password := "password"

	hashed, err := encryption.HashPassword(password)
	require.NoError(t, err)

	err = encryption.CheckPasswordHash(password, hashed)
	require.NoError(t, err)
}
