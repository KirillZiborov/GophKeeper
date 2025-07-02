package encryption

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates bcrypt-hash for a given password.
// Uses bcrypt.DefaultCost for calculation.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash compares a password with its bcrypt-hash.
// Returns nil if the password is correct, or an error otherwise.
func CheckPasswordHash(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
