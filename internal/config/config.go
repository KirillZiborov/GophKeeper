// Package config provides functionalities to parse and manage application configuration.
// It loads configuration settings from environment variables and configuration file.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config contains all app settings.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Security SecurityConfig `mapstructure:"security"`
}

// ServerConfig contains server configuration.
type ServerConfig struct {
	Address string `mapstructure:"address"`
}

// StorageConfig contains storage info.
type StorageConfig struct {
	ConnectionString string `mapstructure:"connection_string"`
}

// SecurityConfig contains security info.
type SecurityConfig struct {
	// JWTKey is the secret key used to sign JWT tokens.
	JWTKey string `mapstructure:"jwt_key"`

	// ExpirationTime specifies the duration for which a JWT token is valid.
	ExpirationTime string `mapstructure:"expiration_time"`
}

// NewConfig loads configuration info from file or environment variables.
// Priority:
// 1) Environment variables (with GOPHKEEPER_ prefix),
// 2) Configuration file (set via server_config.yaml in current directory),
// 3) Default values.
func NewConfig() (*Config, error) {
	// Set default configuration values.
	viper.SetDefault("server.address", "localhost:8080")
	viper.SetDefault("storage.connection_string", "postgres://gophkeeper:1234@localhost:5432/gophkeeper?sslmode=disable")
	viper.SetDefault("security.jwt_key", "supersecretkey")
	viper.SetDefault("security.expiration_time", "3h")

	// Extract environment variables.
	viper.SetEnvPrefix("GOPHKEEPER")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Search for "server_config.yaml".
	viper.AddConfigPath(".")
	viper.AddConfigPath("./internal/config")
	viper.SetConfigName("server_config")

	// Read config from file if found.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("%v\n", err)
	} else {
		fmt.Println("Using config file: ", viper.ConfigFileUsed())
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
