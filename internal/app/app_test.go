// Package app_test provides tests for GophKeeper business logic.
package app_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"

	"github.com/KirillZiborov/GophKeeper/internal/app"
	"github.com/KirillZiborov/GophKeeper/internal/auth"
	"github.com/KirillZiborov/GophKeeper/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateStr() string {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
}

// Test case: Login with random credentials.
func TestLoginWrongCredentials(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := &app.KeeperService{
		Store: fakeStore,
	}

	auth.SetTokenConfig("testsecret", "1h")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Login with random username.
	username := generateStr()
	password := generateStr()
	_, err := svc.Login(ctx, username, password)
	require.Error(t, err)
	assert.Equal(t, "user not found", err.Error())
}

// Test case: Login with wrong password.
func TestWrongPassword(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := &app.KeeperService{
		Store: fakeStore,
	}

	auth.SetTokenConfig("testsecret", "1h")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// User registers.
	username := "user"
	password := "correct"
	token, err := svc.Register(ctx, username, password)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, token, "Token should not be empty")

	// Check that a user is saved in the storage.
	user, err := fakeStore.GetUser(username)
	assert.NoError(t, err, "User should be found after registration")
	assert.Equal(t, username, user.Username)

	// User tries to login with wrong password.
	password = "wrong"
	_, err = svc.Login(ctx, username, password)
	// Expect error.
	require.Error(t, err)
	assert.Equal(t, "invalid username or password", err.Error())
}

// Test of GophKeeper business logic correctness.
func TestKeeperServiceIntegration(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := &app.KeeperService{
		Store: fakeStore,
	}

	auth.SetTokenConfig("testsecret", "1h")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Register test.
	username := generateStr()
	password := "securePassword"
	token, err := svc.Register(ctx, username, password)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, token, "Token should not be empty")

	// Check that a user is saved in the database.
	user, err := fakeStore.GetUser(username)
	assert.NoError(t, err, "User should be found after registration")
	assert.Equal(t, username, user.Username)

	// Login test.
	token, err = svc.Login(ctx, username, password)
	assert.NoError(t, err, "Login should succeed with correct credentials")
	assert.NotEmpty(t, token, "Token should not be empty")

	// Test that userID could be extracted from returned token.
	userID := auth.GetUserID(token)
	assert.NotEmpty(t, userID, "userID should not be empty")

	// AddSecret test.
	data := "encrypted_data_example"
	meta := "some metadata"
	id, err := svc.AddSecret(ctx, userID, data, meta)
	assert.NoError(t, err, "AddSecret should succeed")

	// Get credentials and check that it is saved in the storage.
	creds, err := svc.GetSecrets(ctx, userID)
	assert.NoError(t, err, "GetSecret should succeed")
	assert.GreaterOrEqual(t, len(creds), 1, "There should be at least one secret")

	// EditSecret test.
	newData := "updated_encrypted_data"
	newMeta := "updated metadata"
	err = svc.EditSecret(ctx, id, userID, newData, newMeta)
	assert.NoError(t, err, "EditSecret should succeed")

	// Check that the secret is updated successfully.
	creds, err = svc.GetSecrets(ctx, userID)
	assert.NoError(t, err, "GetSecret after edit should succeed")
	assert.GreaterOrEqual(t, len(creds), 1, "There should be at least one secret")
	assert.Equal(t, newData, creds[0].Data, "Secret data should be updated")
	assert.Equal(t, newMeta, creds[0].Meta, "Secret meta should be updated")
}

// Test case: user tries to update others user secret.
func TestUpdateSecretAccessDenied(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := &app.KeeperService{
		Store: fakeStore,
	}

	auth.SetTokenConfig("testsecret", "1h")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// user1 register test.
	username := "user1"
	password := "securePassword"
	token, err := svc.Register(ctx, username, password)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, token, "Token should not be empty")
	userID := auth.GetUserID(token)
	assert.NotEmpty(t, userID, "userID should not be empty")

	// user1 adds a secret with some generated id.
	data := "encrypted_data_example"
	meta := "some metadata"
	id, err := svc.AddSecret(ctx, userID, data, meta)
	assert.NoError(t, err, "AddSecret should succeed")

	// user2 try to update user1's secret by id.
	err = svc.EditSecret(context.Background(), id, "user2", "new data", "new meta")
	// Expect error.
	require.Error(t, err)
	assert.Equal(t, "access denied: secret doesn't belong to user", err.Error())
}

func TestDoubleRegister(t *testing.T) {
	fakeStore := storage.NewFakeStorage()

	svc := &app.KeeperService{
		Store: fakeStore,
	}

	auth.SetTokenConfig("testsecret", "1h")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Register test for the fisrt time.
	username := "user"
	password := "securePassword"
	token, err := svc.Register(ctx, username, password)
	assert.NoError(t, err, "Registration should succeed")
	assert.NotEmpty(t, token, "Token should not be empty")

	// Register second time with the same username.
	username = "user"
	password = "1234"
	_, err = svc.Register(ctx, username, password)
	// Expect error.
	require.Error(t, err)
	assert.Equal(t, "user already exists", err.Error())
}
