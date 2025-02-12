package grpcapi

import (
	"context"

	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/interceptors"
	"github.com/KirillZiborov/GophKeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetSecret is the gRPC method returning all saved secret data for an authentificated user.
func (s *GophKeeperServer) GetSecret(ctx context.Context, req *proto.GetSecretRequest) (*proto.GetSecretResponse, error) {
	// Extract userID from context set by interceptor.
	userID, ok := interceptors.GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthenticated: no valid token")
	}

	// Call to business logic.
	credsList, err := s.svc.GetSecret(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get Secret: %v", err)
	}

	// Prepare response.
	var protoCreds []*proto.CountedSecret
	for _, c := range credsList {
		protoCreds = append(protoCreds, &proto.CountedSecret{
			Id: c.ID,
			Secret: &proto.Secret{
				Data: c.Data,
				Meta: c.Meta,
			},
		})
	}

	return &proto.GetSecretResponse{
		Secret: protoCreds,
	}, nil
}
