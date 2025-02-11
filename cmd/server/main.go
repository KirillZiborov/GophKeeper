// Package main implements a URL shortener server.
// It initializes configuration, logging and storage (file or database),
// sets up HTTP routes with middleware, registers pprof handlers for profiling,
// and starts the HTTP server.
package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KirillZiborov/GophKeeper/internal/app"
	"github.com/KirillZiborov/GophKeeper/internal/config"
	"github.com/KirillZiborov/GophKeeper/internal/grpcapi"
	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/interceptors"
	"github.com/KirillZiborov/GophKeeper/internal/grpcapi/proto"
	"github.com/KirillZiborov/GophKeeper/internal/logging"
	"github.com/KirillZiborov/GophKeeper/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	db    *pgxpool.Pool
	store storage.Storage

	// Use go run -ldflags to set up build variables while compiling.
	buildVersion = "N/A" // Build version
	buildDate    = "N/A" // Build date
	buildCommit  = "N/A" // Build commit
)

// main is the entrypoint of the GophKeeper server.
// It initializes configuration, logging and storage, sets up HTTP routes with middleware,
// registers pprof handlers for profiling, and starts the HTTP server.
func main() {
	// Print build info.
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	// Initialize the logging system.
	err := logging.Initialize()
	if err != nil {
		logging.Sugar.Errorw("Internal logging error", "error", err)
	}

	// Load the configuration.
	cfg := config.NewConfig()

	// Initialize storage based on the configuration.
	if cfg.DBPath != "" {
		// Establish a connection to the PostgreSQL database with a timeout.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db, err = pgxpool.New(ctx, cfg.DBPath)
		if err != nil {
			logging.Sugar.Errorw("Unable to connect to database", "error", err)
			return
		}

		// Create users and secrets table in the database if it doesn't exist.
		err = storage.CreateTables(ctx, db)
		if err != nil {
			logging.Sugar.Errorw("Failed to create table", "error", err)
			return
		}
		defer db.Close()

		// Use the database store for URL storage.
		store = storage.NewDBStore(db)
	} else {
		// If no database is configured, use a file-based store.
		logging.Sugar.Errorw("No database path specified", "error", err)
		return
	}

	service := app.KeeperService{
		Store: store,
		Cfg:   cfg,
	}

	// Start the gRPC server.
	lis, err := net.Listen("tcp", cfg.GRPCAddress)
	if err != nil {
		logging.Sugar.Fatalw("Failed to listen on gRPC address", "error", err)
	}

	grpcServer := grpc.NewServer(
		// Add authentificatrion interceptor.
		grpc.ChainUnaryInterceptor(interceptors.AuthInterceptor()),
	)
	// Register the gRPC service.
	proto.RegisterKeeperServer(grpcServer, grpcapi.NewGRPCKeeperServer(&service))
	// Register reflection for grpcurl.
	reflection.Register(grpcServer)

	// Start gRPC server in goroutine.
	go func() {
		logging.Sugar.Infow("Starting gRPC server", "address", cfg.GRPCAddress)
		if err := grpcServer.Serve(lis); err != nil {
			logging.Sugar.Errorw("gRPC server exited with error", "error", err)
			return
		}
	}()

	// Handle sys calls for graceful shutdown.
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	sig := <-stopChan
	logging.Sugar.Infow("Received shutdown signal", "signal", sig)

	_, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	grpcServer.GracefulStop()
	logging.Sugar.Infow("Server shutdown complete")
}
