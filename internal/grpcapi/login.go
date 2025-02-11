package grpcapi

import (
	"context"
	"fmt"

	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Login аутентифицирует пользователя по username и password.
// При успешном логине сервер генерирует JWT-токен и передаёт его через header.
func (s *GophKeeperServer) Login(ctx context.Context, req *proto.LoginRequest) (*proto.LoginResponse, error) {
	userData := req.GetUserData()
	if userData == nil || userData.Username == "" || userData.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password must be provided")
	}

	token, err := s.svc.Login(ctx, userData.Username, userData.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	// Устанавливаем токен в header ответа.
	if err := setResponseToken(ctx, token); err != nil {
		fmt.Printf("Warning: failed to set response token: %v\n", err)
	}

	return &proto.LoginResponse{}, nil
}
