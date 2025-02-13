package grpcapi

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// setResponseToken sets token in response header using grpc.SetHeader.
func setResponseToken(ctx context.Context, token string) error {
	md := metadata.Pairs("token", token)
	return grpc.SetHeader(ctx, md)
}
