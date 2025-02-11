// Package auth provides authentication utilities using JSON Web Tokens (JWT).
package auth

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// Claims defines the structure of JWT claims used in the authentication process.
type Claims struct {
	jwt.RegisteredClaims        // Standart JWT fields.
	UserID               string //UserID - unique ID of the user.
}

// TokenExp specifies the duration for which a JWT token is valid.
// Tokens expire 3 hours after issuance.
const TokenExp = time.Hour * 3

// SecretKey is the secret key used to sign JWT tokens.
const SecretKey = "supersecretkey"

// GenerateToken creates a new JWT token for a given userID.
// If the provided userID is empty, it generates a new UUID for the user.
// The function returns the signed JWT token string or an error if the process fails.
func GenerateToken(userID string) (string, error) {
	if userID == "" {
		userID = uuid.New().String()
	}

	tokenString, err := BuildJWTString(userID)
	if err != nil {
		log.Fatal(err)
	}

	return tokenString, nil
}

// BuildJWTString creates a signed JWT token string for a given userID.
// It sets the token's expiration time based on the TokenExp constant.
// The function uses the HS256 signing method and returns the signed token string or an error.
func BuildJWTString(userID string) (string, error) {
	// Create a new token with given claims.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},

		UserID: userID,
	})

	// Sign token using SecretKey.
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserID extracts the UserID from a given JWT token string.
// It parses the token, validates its signature and expiration, and retrieves the UserID claim.
// If the token is invalid or expired, the function returns an empty string.
func GetUserID(tokenString string) string {
	claims := &Claims{}
	// Parse token and extract claims.
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		return ""
	}

	if !token.Valid {
		fmt.Println("Token is not valid")
		return ""
	}

	// Debug: fmt.Println("Token is valid")
	return claims.UserID
}
