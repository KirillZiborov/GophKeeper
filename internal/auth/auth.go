// Package auth provides authentication utilities using JSON Web Tokens (JWT).
// It also provides gRPC interceptor for handling authentication.
package auth

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Claims defines the structure of JWT claims used in the authentication process.
type Claims struct {
	jwt.RegisteredClaims        // Standart JWT fields.
	UserID               string //UserID - unique ID of the user.
}

// contextKey defines a type of a key for storing user ID value in context.
// Defined to avoid staticcheck warnings.
type contextKey string

const (
	// metadataKey is the key in context where we store userID.
	metadataKey contextKey = "userID"

	// cookieHeader is the name of metadata that simulates cookie.
	cookieHeader = "token"
)

// JWToken holds the configuration for token generation.
type JWToken struct {
	SecretKey string
	TokenExp  time.Duration
}

var tokenConfig JWToken

// SetTokenConfig sets JWT parameters from the configuration.
func SetTokenConfig(secret string, exp string) {
	tokenConfig.SecretKey = secret

	expTime, err := time.ParseDuration(exp)
	if err != nil {
		log.Printf("Bad token expiration configuration: %v, using default", err)
		expTime = 3 * time.Hour
	}

	tokenConfig.TokenExp = expTime
}

// AuthInterceptor is a gRPC interceptor that handles users authentification.
// It emulates HTTP cookie-based JWT app.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentification for register and login.
		if info.FullMethod == "/proto.Keeper/Login" || info.FullMethod == "/proto.Keeper/Register" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md.Get(cookieHeader)) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "token is required")
		}

		token := md.Get(cookieHeader)[0]

		// Parse and validate cookie from metadata.
		userID := GetUserID(token)
		if userID == "" {
			return nil, status.Errorf(codes.Unauthenticated, "Invalid token in %s", cookieHeader)
		}

		// Put userID in context.
		newCtx := context.WithValue(ctx, metadataKey, userID)

		// Call next handler.
		resp, err := handler(newCtx, req)
		return resp, err
	}
}

// GenerateToken creates a new JWT token for a given userID.
// If the provided userID is empty, it generates a new UUID for the user.
// The function returns the signed JWT token string or an error if the process fails.
func GenerateToken(userID string) (string, error) {
	if userID == "" {
		fmt.Println("Warning: no userID found")
		userID = uuid.New().String()
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenConfig.TokenExp)),
		},
		UserID: userID,
	})
	tokenString, err := token.SignedString([]byte(tokenConfig.SecretKey))
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
		return []byte(tokenConfig.SecretKey), nil
	})
	if err != nil {
		return ""
	}

	if !token.Valid {
		fmt.Println("Token is not valid")
		return ""
	}

	return claims.UserID
}

// GetUserIDFromContext extracts userID from context in gRPC methods.
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(metadataKey)
	userID, ok := val.(string)
	return userID, ok
}
