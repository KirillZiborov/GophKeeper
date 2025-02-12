package grpcapi

import (
	"context"

	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/interceptors"
	"github.com/KirillZiborov/GophKeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AddSecret is the gRPC method for adding secret data for an authentificated user.
// Authentification is performed via interceptor (extracting token from metadata).
func (s *GophKeeperServer) AddSecret(ctx context.Context, req *proto.AddSecretRequest) (*proto.AddSecretResponse, error) {
	// Extract userID from context set by interceptor.
	userID, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated: no valid token")
	}

	creds := req.GetSecret()
	if creds == nil || creds.Data == "" {
		return nil, status.Error(codes.InvalidArgument, "Secret data must be provided")
	}

	// Call to business logic.
	id, err := s.svc.AddSecret(ctx, userID, creds.Data, creds.Meta)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to add Secret: %v", err)
	}

	return &proto.AddSecretResponse{Id: id}, nil
}
