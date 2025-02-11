package grpcapi

import (
	"context"
	"fmt"

	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register is the gRPC ...
// Register регистрирует нового пользователя.
// Клиент отправляет RegisterRequest с данными пользователя (username, password).
// В ответ сервер генерирует JWT-токен и возвращает его через заголовок ответа (так как proto-структура RegisterResponse пуста).
func (s *GophKeeperServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
	userData := req.GetUserData()
	if userData == nil || userData.Username == "" || userData.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password must be provided")
	}

	// Вызываем бизнес-логику регистрации.
	token, err := s.svc.Register(ctx, userData.Username, userData.Password)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "registration failed: %v", err)
	}

	// Возвращаем пустой ответ, но устанавливаем токен в header, чтобы клиент смог его сохранить.
	if err := setResponseToken(ctx, token); err != nil {
		// Если не удалось установить заголовок, логируем, но не блокируем выполнение.
		fmt.Printf("Warning: failed to set response token: %v\n", err)
	}

	return &proto.RegisterResponse{}, nil
}
