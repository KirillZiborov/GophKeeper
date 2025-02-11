// Package config provides functionalities to parse and manage application configuration.
// It loads configuration settings from environment variables, command-line flags
// and configuration file.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/KirillZiborov/GophKeeper/internal/logging"
)

// Config represents the configuration settings for the application.
// It includes settings for the server address, base URL for shortened URLs,
// file storage path, database connection string and trusted subnet CIDR.
type Config struct {
	// DBPath contains the database connection string used to connect
	// to the PostgreSQL database. If empty, the application uses file storage.
	// Example: "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	DBPath string `json:"database_dsn"`
	// GRPCAddress specifies the address on which the gRPC server listens.
	// Example: "localhost:8080".
	// If empty, gRPC is disabled.
	GRPCAddress string `json:"grpc_address"`
}

// NewConfig initializes and returns a new coniguration instance.
// It parses command-line flags and overrides them with environment variables if they are set.
// The priority is:
// 1. Environment Variables
// 2. Command-Line Flags
// 3. Configuration File
// 4. Default Values
//
// 1. Environment Variables:
//
//	DATABASE_DSN         Overrides the -d flag.
//	GRPC_ADDRESS       	 Overrides the -a flag.
//
// 2. Command-Line Flags:
//
//	-d string
//	      Database address
//	-a string
//	      Address of the gRPC server
//	-config string
//	      Configuration file path
//
// 3. Configuration File:
//
//	"database_dsn": string
//		  Analogue for environment variable DATABASE_DSN and -d flag
//	"grpc_address": string
//		  Analogue for environment variable GRPC_ADDRESS and -g flag
//
// 4. Default Values:
//
//	DBPath:      	"postgres://gophkeeper:1234@localhost:5432/gophkeeper?sslmode=disable",
//	GRPCAddress:    "localhost:8080"
func NewConfig() *Config {
	cfg := &Config{}
	// Specify default configuration values.
	currentCfg := &Config{
		DBPath:      "postgres://gophkeeper:1234@localhost:5432/gophkeeper?sslmode=disable",
		GRPCAddress: "localhost:8080",
	}

	// Define command-line flags and associate them with Config fields.
	flag.StringVar(&cfg.DBPath, "d", "", "Database address")
	flag.StringVar(&cfg.GRPCAddress, "a", "", "Address of the gRPC server")

	var configPath string
	flag.StringVar(&configPath, "config", "", "Path to configuration file")

	// Parse the command-line flags.
	flag.Parse()

	// Check if configuration is set in file.
	if configFilePath := getConfigFilePath(configPath); configFilePath != "" {
		// Load configuration from the file and override default values.
		if err := loadConfigFromFile(configFilePath, currentCfg); err != nil {
			logging.Sugar.Errorw("Failed to load config file at", "error", err, "addr", configFilePath)
		}
	}

	// Override DBPath with the DATABASE_DSN environment variable if set.
	if DBPath := os.Getenv("DATABASE_DSN"); DBPath != "" {
		cfg.DBPath = DBPath
	} else if cfg.DBPath == "" {
		// Look for DBPath in the config file.
		cfg.DBPath = currentCfg.DBPath
	}

	// Override GRPCAddress with the GRPC_ADDRESS environment variable if set.
	grpcAddress := os.Getenv("GRPC_ADDRESS")
	if grpcAddress != "" {
		cfg.GRPCAddress = grpcAddress
	} else if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = currentCfg.GRPCAddress
	}

	return cfg
}

// getConfigFilePath returns the path to the configuration file if exists.
func getConfigFilePath(flagPath string) string {
	if envPath := os.Getenv("CONFIG"); envPath != "" {
		return envPath
	} else if flagPath != "" {
		return flagPath
	}
	log.Println("No configuration file path provided.")
	return ""
}

// loadConfigFromFile reads the configuration from a JSON file.
func loadConfigFromFile(filePath string, cfg *Config) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(cfg); err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}
	return nil
}
