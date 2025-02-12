// Package grpcapi provides functionality for handling gRPC communication with the GophKeeper service.
// It includes interfaces and structs for defining gRPC services and servers, as well as methods
// for interacting with the GophKeeper service.
package grpcapi

import (
	"github.com/KirillZiborov/GophKeeper/internal/app"
	"github.com/KirillZiborov/GophKeeper/proto"
)

// GophKeeperServer supports all neccessary server methods.
type GophKeeperServer struct {
	proto.UnimplementedKeeperServer
	svc *app.KeeperService
}

// NewGRPCKeeperServer creates a new instance of the GophKeeperServer struct with the provided service.
// It initializes the service field of the GophKeeperServer struct with the given
// service instance and returns a pointer to the newly created GophKeeperServer instance.
func NewGRPCKeeperServer(svc *app.KeeperService) *GophKeeperServer {
	return &GophKeeperServer{svc: svc}
}
