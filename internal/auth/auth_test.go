package auth_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/KirillZiborov/GophKeeper/internal/auth"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestGenerateTokenAndExtractUserID(t *testing.T) {
	// Устанавливаем конфигурацию токена.
	auth.SetTokenConfig("test-secret", "2h")

	userID := "user"
	token, err := auth.GenerateToken(userID)
	require.NoError(t, err)

	extractedUserID := auth.GetUserID(token)
	require.Equal(t, userID, extractedUserID, "Extracted userID should match the original")
}

func TestAuthInterceptor_GetUserIDFromContext(t *testing.T) {
	auth.SetTokenConfig("test-secret", "2h")

	token, err := auth.GenerateToken("user1")
	require.NoError(t, err)

	md := metadata.Pairs("token", token)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		uid, ok := auth.GetUserIDFromContext(ctx)
		if !ok {
			return nil, fmt.Errorf("no userID in context")
		}
		return uid, nil
	}

	interceptor := auth.AuthInterceptor()
	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/proto.Keeper/SomeMethod"}, handler)
	require.NoError(t, err)
	require.Equal(t, "user1", resp)
}

func TestAuthInterceptor_NoToken(t *testing.T) {
	auth.SetTokenConfig("test-secret", "2h")

	ctx := context.Background()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}

	interceptor := auth.AuthInterceptor()
	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/proto.Keeper/SomeMethod"}, handler)
	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, st.Code(), status.Code(err))
}
