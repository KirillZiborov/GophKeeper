package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("ENV vars only", func(t *testing.T) {
		os.Setenv("GRPC_ADDRESS", "localhost:1717")
		os.Setenv("DATABASE_DSN", "postgres://database")

		cfg := NewConfig()

		assert.Equal(t, "localhost:1717", cfg.GRPCAddress)
		assert.Equal(t, "postgres://database", cfg.DBPath)

		os.Unsetenv("GRPC_ADDRESS")
		os.Unsetenv("DATABASE_DSN")
	})

	t.Run("ENV vars + flags", func(t *testing.T) {
		// anti-panic: flag redefined
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		os.Setenv("GRPC_ADDRESS", "localhost:1717")
		os.Setenv("DATABASE_DSN", "postgres://database")
		os.Args = []string{"program", "-a", "localhost:2717", "-d", "postgres://fdatabase"}

		cfg := NewConfig()

		assert.Equal(t, "localhost:1717", cfg.GRPCAddress)
		assert.Equal(t, "postgres://database", cfg.DBPath)

		os.Unsetenv("GRPC_ADDRESS")
		os.Unsetenv("DATABASE_DSN")
	})

	t.Run("flags only", func(t *testing.T) {
		// anti-panic: flag redefined
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		os.Args = []string{"program", "-a", "localhost:2717", "-d", "postgres://fdatabase"}

		cfg := NewConfig()

		assert.Equal(t, "localhost:2717", cfg.GRPCAddress)
		assert.Equal(t, "postgres://fdatabase", cfg.DBPath)
	})

	t.Run("config file only", func(t *testing.T) {
		// anti-panic: flag redefined
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		configFile := `{
				"grpc_address": "localhost:3717",
    			"database_dsn": "postgres://database"
			}`

		filename := "config.json"
		tempFile, err := os.Create(filename)
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString(configFile)
		require.NoError(t, err)

		os.Args = []string{"program", "-config", filename}

		cfg := NewConfig()

		assert.Equal(t, "localhost:3717", cfg.GRPCAddress)
		assert.Equal(t, "postgres://database", cfg.DBPath)
	})

	t.Run("ENV vars + config file", func(t *testing.T) {
		// anti-panic: flag redefined
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		os.Setenv("GRPC_ADDRESS", "localhost:1717")
		os.Setenv("DATABASE_DSN", "postgres://database")

		configFile := `{
				"grpc_address": "localhost:3717",
    			"database_dsn": "postgres://cdatabase"
			}`

		tempFile, err := os.CreateTemp("", "config.json")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString(configFile)
		require.NoError(t, err)

		os.Args = []string{"program", "-config", tempFile.Name()}

		cfg := NewConfig()

		// ENV vars should take priority over config file
		assert.Equal(t, "localhost:1717", cfg.GRPCAddress)
		assert.Equal(t, "postgres://database", cfg.DBPath)

		os.Unsetenv("GRPC_ADDRESS")
		os.Unsetenv("DATABASE_DSN")
	})

	t.Run("config file + flags", func(t *testing.T) {
		// anti-panic: flag redefined
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		configFile := `{
				"grpc_address": "localhost:3717",
    			"database_dsn": "postgres://cdatabase"
			}`

		tempFile, err := os.CreateTemp("", "config.json")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString(configFile)
		require.NoError(t, err)

		os.Args = []string{"program", "-a", "localhost:2717", "-d", "postgres://fdatabase"}

		cfg := NewConfig()

		// Flags should take priority over config file
		assert.Equal(t, "localhost:2717", cfg.GRPCAddress)
		assert.Equal(t, "postgres://fdatabase", cfg.DBPath)
	})

	t.Run("ENV vars + config file + flags", func(t *testing.T) {
		// anti-panic: flag redefined
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		os.Setenv("GRPC_ADDRESS", "localhost:1717")
		os.Setenv("DATABASE_DSN", "postgres://database")

		configFile := `{
				"grpc_address": "localhost:3717",
    			"database_dsn": "postgres://cdatabase"
			}`

		tempFile, err := os.CreateTemp("", "config.json")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		_, err = tempFile.WriteString(configFile)
		require.NoError(t, err)

		os.Args = []string{"program", "-a", "localhost:2717", "-d", "postgres://fdatabase"}

		cfg := NewConfig()

		// ENV vars should take priority over flags and config file
		assert.Equal(t, "localhost:1717", cfg.GRPCAddress)
		assert.Equal(t, "postgres://database", cfg.DBPath)

		os.Unsetenv("GRPC_ADDRESS")
		os.Unsetenv("DATABASE_DSN")
	})

	t.Run("default values", func(t *testing.T) {
		// anti-panic: flag redefined
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
		// no flags
		os.Args = []string{"program"}

		cfg := NewConfig()

		assert.Equal(t, "localhost:8080", cfg.GRPCAddress)
		assert.Equal(t, "postgres://postgres:12345678@localhost:5432/gophkeeper", cfg.DBPath)
	})
}
