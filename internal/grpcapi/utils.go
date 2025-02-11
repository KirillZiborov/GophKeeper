package grpcapi

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// setResponseToken устанавливает токен в заголовки ответа.
// В gRPC для установки заголовков используется grpc.SetHeader.
func setResponseToken(ctx context.Context, token string) error {
	md := metadata.Pairs("token", token)
	return grpc.SetHeader(ctx, md)
}
