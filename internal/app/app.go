// Package app provides the business logic for sign up, login and managing user's secrets.
// It includes functionality to add user and login as well as add and update user's secrets.
package app

import (
	"context"
	"errors"

	"github.com/KirillZiborov/GophKeeper/internal/auth"
	"github.com/KirillZiborov/GophKeeper/internal/config"
	"github.com/KirillZiborov/GophKeeper/internal/models"
	"github.com/KirillZiborov/GophKeeper/internal/storage"
	"github.com/KirillZiborov/GophKeeper/pkg/encryption"
	"github.com/google/uuid"
)

// ErrUserNotFound is login error.
var ErrUserNotFound = errors.New("invalid username or password")

// ErrAccessDenied is returned when user try to approach not his secret.
var ErrAccessDenied = errors.New("access denied: secret doesn't belong to user")

// KeeperService is a facade of GophKeeper business logic.
type KeeperService struct {
	Store storage.Storage // Using database storage.
	Cfg   *config.Config  // Using configuration.
}

// Register adds new user to GophKeeper saving it username and hashed password.
func (ks *KeeperService) Register(ctx context.Context, username, password string) (string, error) {

	userUuid := uuid.New().String()

	// Hash password.
	hashedPassword, err := encryption.HashPassword(password)
	if err != nil {
		return "", err
	}

	// Create user's model.
	user := &models.User{
		ID:       userUuid,
		Username: username,
		Password: hashedPassword,
	}

	// Save user in the database.
	if err = ks.Store.RegisterUser(user); err != nil {
		return "", err
	}

	// Generates token for a new user.
	token, err := auth.GenerateToken(userUuid)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Login authentificates user with username and password provided.
func (ks *KeeperService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := ks.Store.GetUser(username)
	if err != nil {
		return "", err
	}

	if err := encryption.CheckPasswordHash(password, user.Password); err != nil {
		return "", ErrUserNotFound
	}

	// Generates token for user.
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// AddSecret adds secret data to user's list of credentials.
// Parameter token is a JWT which is used for etracting userID.
func (ks *KeeperService) AddSecret(ctx context.Context, userID, data, meta string) (int64, error) {
	creds := &models.Secret{
		Data:   data,
		Meta:   meta,
		UserID: userID,
	}

	return ks.Store.AddSecret(creds)
}

// EditSecret updates secret data using its id.
func (ks *KeeperService) EditSecret(ctx context.Context, id int64, userID, data, meta string) error {
	secret, err := ks.Store.GetSecretByID(id)
	if err != nil {
		return err
	}

	if secret.UserID != userID {
		return ErrAccessDenied
	}

	secret.Data = data
	secret.Meta = meta

	return ks.Store.EditSecret(secret)
}

// GetSecret retrieves all user's credentials.
// Parameter token is a JWT which is used for etracting userID.
func (ks *KeeperService) GetSecrets(ctx context.Context, userID string) ([]models.Secret, error) {
	// Retrieve all user's secrets from storage.
	creds, err := ks.Store.GetSecrets(userID)
	if err != nil {
		return nil, err
	}
	return creds, nil
}
