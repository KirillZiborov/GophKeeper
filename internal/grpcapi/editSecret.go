package grpcapi

import (
	"context"

	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/interceptors"
	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EditSecret редактирует приватные данные по их идентификатору для аутентифицированного пользователя.
func (s *GophKeeperServer) EditSecret(ctx context.Context, req *proto.EditSecretRequest) (*proto.EditSecretResponse, error) {
	userID, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated: no valid token")
	}

	// Преобразуем идентификатор из строки в целое число.
	id := req.GetId()
	if id == "" {
		return nil, status.Errorf(codes.InvalidArgument, "id field is empty")
	}

	creds := req.GetSecret()
	if creds == nil || creds.Data == "" {
		return nil, status.Error(codes.InvalidArgument, "Secret data must be provided")
	}

	err := s.svc.EditSecret(ctx /* token не передается, используем interceptor */, "", id, creds.Data, creds.Meta)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to edit Secret: %v", err)
	}

	return &proto.EditSecretResponse{}, nil
}
