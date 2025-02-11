package grpcapi

import (
	"context"

	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/interceptors"
	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AddSecret добавляет приватные данные для аутентифицированного пользователя.
// Аутентификация выполняется через interceptor (извлечение токена из metadata).
func (s *GophKeeperServer) AddSecret(ctx context.Context, req *proto.AddSecretRequest) (*proto.AddSecretResponse, error) {
	// Извлекаем userID из контекста, установленного interceptor-ом.
	userID, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated: no valid token")
	}

	creds := req.GetSecret()
	if creds == nil || creds.Data == "" {
		return nil, status.Error(codes.InvalidArgument, "Secret data must be provided")
	}

	err := s.svc.AddSecret(ctx /* token не передается в теле, а используется из interceptor */, "", creds.Data, creds.Meta)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add Secret: %v", err)
	}

	return &proto.AddSecretResponse{}, nil
}
