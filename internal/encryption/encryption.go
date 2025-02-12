package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptWithKey encodes plaintext using AES-GCM and encryption key.
// encryptionKey hashed using SHA256 in order to obtain 32 bytes AES-256 key.
func EncryptWithKey(plaintext, encryptionKey string) (string, error) {
	if encryptionKey == "" {
		return "", errors.New("encryption key is empty")
	}
	// Get 32 bytes key using SHA256.
	hasher := sha256.New()
	hasher.Write([]byte(encryptionKey))
	key := hasher.Sum(nil)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return encoded, nil
}

// DecryptWithKey decodes encoded text using AES-GCM and encryption key.
// encryptionKey hashed using SHA256 in order to obtain 32 bytes AES-256 key.
func DecryptWithKey(encodedCiphertext string, encryptionKey string) (string, error) {
	if encryptionKey == "" {
		return "", errors.New("encryption key is empty")
	}

	// Get 32 bytes key using SHA256.
	hasher := sha256.New()
	hasher.Write([]byte(encryptionKey))
	key := hasher.Sum(nil)

	// Decode Base64 to []byte.
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and encoded data.
	nonce := ciphertext[:nonceSize]
	data := ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
