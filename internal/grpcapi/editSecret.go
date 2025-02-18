package grpcapi

import (
	"context"

	"github.com/KirillZiborov/GophKeeper/internal/auth"
	"github.com/KirillZiborov/GophKeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EditSecret is the gRPC method for updating secret data by id for an authentificated user.
func (s *GophKeeperServer) EditSecret(ctx context.Context, req *proto.EditSecretRequest) (*proto.EditSecretResponse, error) {
	// Extract userID from context set by interceptor.
	userID, ok := auth.GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated: no valid token")
	}

	id := req.GetId()

	secret := req.GetSecret()
	if secret == nil || secret.Data == "" {
		return nil, status.Error(codes.InvalidArgument, "Secret data must be provided")
	}

	// Call to business logic.
	err := s.svc.EditSecret(ctx, id, userID, secret.Data, secret.Meta)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to edit Secret: %v", err)
	}

	return &proto.EditSecretResponse{}, nil
}
