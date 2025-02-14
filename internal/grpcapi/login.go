package grpcapi

import (
	"context"
	"fmt"

	"github.com/KirillZiborov/GophKeeper/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Login is the gRPC method for user authentification with provided username and password.
// If successfull, generates JWT token and sets it to the response header.
func (s *GophKeeperServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	userData := req.GetUserData()
	if userData == nil || userData.Username == "" || userData.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password must be provided")
	}

	// Call to business logic.
	token, err := s.svc.Login(ctx, userData.Username, userData.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	// Set token to the response header.
	if err := setResponseToken(ctx, token); err != nil {
		fmt.Printf("Warning: failed to set response token: %v\n", err)
	}

	return &proto.LoginResponse{}, nil
}
