package grpcapi

import (
	"context"
	"fmt"

	"github.com/KirillZiborov/GophKeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register is the gRPC method for adding a new user to the service.
// Client sets RegisterRequest with username and password.
// In response, server generates JWT token and sets it to the response header.
func (s *GophKeeperServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	userData := req.GetUserData()
	if userData == nil || userData.Username == "" || userData.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password must be provided")
	}

	// Call to business logic.
	token, err := s.svc.Register(ctx, userData.Username, userData.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	// Set token to the response header.
	if err := setResponseToken(ctx, token); err != nil {
		fmt.Printf("Warning: failed to set response token: %v\n", err)
	}

	return &proto.RegisterResponse{}, nil
}
