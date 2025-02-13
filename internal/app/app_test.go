package app_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/KirillZiborov/GophKeeper/internal/app"
	"github.com/KirillZiborov/GophKeeper/internal/auth"
	"github.com/KirillZiborov/GophKeeper/internal/config"
	"github.com/KirillZiborov/GophKeeper/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
)

func generateStr() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

func TestKeeperServiceIntegration(t *testing.T) {
	dsn := "postgres://gophkeeper:1234@localhost:5432/test_gophkeeper?sslmode=disable"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	assert.NoError(t, err, "Failed to connect to PostgreSQL")
	defer pool.Close()

	err = storage.CreateTables(ctx, pool)
	assert.NoError(t, err, "Failed to create tables")

	store := storage.NewDBStore(pool)

	cfg := &config.Config{
		Server:  config.ServerConfig{Address: "http://localhost:8080"},
		Storage: config.StorageConfig{ConnectionString: dsn},
	}

	svc := &app.KeeperService{
		Store: store,
		Cfg:   cfg,
	}

	auth.SetTokenConfig("testsecret", "1h")

	// Register test.
	username := generateStr()
	password := "securePassword"
	token, err := svc.Register(ctx, username, password)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, token, "Token should not be empty")

	// Check that a user is saved in the database.
	user, err := store.GetUser(username)
	assert.NoError(t, err, "User should be found after registration")
	assert.Equal(t, username, user.Username)

	// Login test.
	token, err = svc.Login(ctx, username, password)
	userID := auth.GetUserID(token)
	assert.NoError(t, err, "Login should succeed with correct Secret")
	assert.NotEmpty(t, token, "Token should not be empty")
	assert.NotEmpty(t, userID, "userID should not be empty")

	// AddSecret test.
	data := "encrypted_data_example"
	meta := "some metadata"
	_, err = svc.AddSecret(ctx, userID, data, meta)
	assert.NoError(t, err, "AddSecret should succeed")

	// Get credentials and check that it is saved in the database.
	creds, err := svc.GetSecret(ctx, userID)
	assert.NoError(t, err, "GetSecret should succeed")
	assert.GreaterOrEqual(t, len(creds), 1, "There should be at least one Secret")

	// EditSecret test.
	// Use the first found record from storage.
	credID := creds[0].ID
	newData := "updated_encrypted_data"
	newMeta := "updated metadata"
	err = svc.EditSecret(ctx, credID, userID, newData, newMeta)
	assert.NoError(t, err, "EditSecret should succeed")

	// Check that the secret is updated successfully.
	creds, err = svc.GetSecret(ctx, userID)
	assert.NoError(t, err, "GetSecret after edit should succeed")
	assert.GreaterOrEqual(t, len(creds), 1, "There should be at least one Secret")
	assert.Equal(t, newData, creds[0].Data, "Secret data should be updated")
	assert.Equal(t, newMeta, creds[0].Meta, "Secret meta should be updated")
}
