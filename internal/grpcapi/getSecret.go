package grpcapi

import (
	"context"

	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/interceptors"
	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetSecret возвращает все приватные данные для аутентифицированного пользователя.
func (s *GophKeeperServer) GetSecret(ctx context.Context, req *proto.GetSecretRequest) (*proto.GetSecretResponse, error) {
	userID, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated: no valid token")
	}

	credsList, err := s.svc.GetSecret(ctx /* token не передается, используем interceptor */, "")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get Secret: %v", err)
	}

	// Преобразуем список моделей в proto-структуры.
	var protoCreds []*proto.Secret
	for _, c := range credsList {
		protoCreds = append(protoCreds, &proto.Secret{
			Data: c.Data,
			Meta: c.Meta,
		})
	}

	return &proto.GetSecretResponse{
		Secret: protoCreds,
	}, nil
}
