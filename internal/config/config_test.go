package config_test

import (
	"flag"
	"os"
	"testing"

	"github.com/KirillZiborov/GophKeeper/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ENVVarsOnly(t *testing.T) {
	// Устанавливаем переменные окружения с префиксом GOPHKEEPER_
	os.Setenv("GOPHKEEPER_SERVER_ADDRESS", "localhost:1717")
	os.Setenv("GOPHKEEPER_STORAGE_CONNECTION_STRING", "postgres://database")
	os.Setenv("GOPHKEEPER_SECURITY_JWT_KEY", "env-secret")
	os.Setenv("GOPHKEEPER_SECURITY_EXPIRATION_TIME", "1h")

	cfg, err := config.NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "localhost:1717", cfg.Server.Address)
	assert.Equal(t, "postgres://database", cfg.Storage.ConnectionString)
	assert.Equal(t, "env-secret", cfg.Security.JWTKey)
	assert.Equal(t, "1h", cfg.Security.ExpirationTime)

	os.Unsetenv("GOPHKEEPER_SERVER_ADDRESS")
	os.Unsetenv("GOPHKEEPER_STORAGE_CONNECTION_STRING")
	os.Unsetenv("GOPHKEEPER_SECURITY_JWT_KEY")
	os.Unsetenv("GOPHKEEPER_SECURITY_EXPIRATION_TIME")
}

func TestConfig_DefaultValues(t *testing.T) {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Unsetenv("GOPHKEEPER_SERVER_ADDRESS")
	os.Unsetenv("GOPHKEEPER_STORAGE_CONNECTION_STRING")
	os.Unsetenv("GOPHKEEPER_SECURITY_JWT_KEY")
	os.Unsetenv("GOPHKEEPER_SECURITY_EXPIRATION_TIME")

	cfg, err := config.NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "localhost:8080", cfg.Server.Address)
	assert.Equal(t, "postgres://gophkeeper:1234@localhost:5432/gophkeeper?sslmode=disable", cfg.Storage.ConnectionString)
	assert.Equal(t, "supersecretkey", cfg.Security.JWTKey)
	assert.Equal(t, "3h", cfg.Security.ExpirationTime)
}

func TestConfig_ConfigFileOnly(t *testing.T) {
	configFile := `server:
 address: ":3717"
storage:
 connection_string: "postgres://file_database"
security:
 jwt_key: "file-secret"
 expiration_time: "2h"`
	tempFile, err := os.Create("server_config.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(configFile)
	require.NoError(t, err)
	tempFile.Close()

	cfg, err := config.NewConfig()
	require.NoError(t, err)

	assert.Equal(t, ":3717", cfg.Server.Address)
	assert.Equal(t, "postgres://file_database", cfg.Storage.ConnectionString)
	assert.Equal(t, "file-secret", cfg.Security.JWTKey)
	assert.Equal(t, "2h", cfg.Security.ExpirationTime)
}

func TestConfig_ENVVarsAndConfigFile(t *testing.T) {
	os.Setenv("GOPHKEEPER_SERVER_ADDRESS", "localhost:1717")
	os.Setenv("GOPHKEEPER_STORAGE_CONNECTION_STRING", "postgres://env_database")
	os.Setenv("GOPHKEEPER_SECURITY_JWT_KEY", "env-secret")
	os.Setenv("GOPHKEEPER_SECURITY_EXPIRATION_TIME", "4h")

	configFile := `server:
 address: ":2717"
storage:
 connection_string: "postgres://file_database_"
security:
 jwt_key: "file-secret-2"
 expiration_time: "2h"`
	tempFile, err := os.Create("server_config.yaml")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	_, err = tempFile.WriteString(configFile)
	require.NoError(t, err)
	tempFile.Close()

	cfg, err := config.NewConfig()
	require.NoError(t, err)

	assert.Equal(t, "localhost:1717", cfg.Server.Address)
	assert.Equal(t, "postgres://env_database", cfg.Storage.ConnectionString)
	assert.Equal(t, "env-secret", cfg.Security.JWTKey)
	assert.Equal(t, "4h", cfg.Security.ExpirationTime)

	os.Unsetenv("GOPHKEEPER_SERVER_ADDRESS")
	os.Unsetenv("GOPHKEEPER_STORAGE_CONNECTION_STRING")
	os.Unsetenv("GOPHKEEPER_SECURITY_JWT_KEY")
	os.Unsetenv("GOPHKEEPER_SECURITY_EXPIRATION_TIME")
}
