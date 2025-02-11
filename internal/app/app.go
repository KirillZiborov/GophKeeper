// Package app provides the business logic for sign up, login and managing user's secrets.
// It includes functionality to add user and login as well as add and update user's secrets.
package app

import (
	"context"
	"errors"

	"github.com/KirillZiborov/GophKeeper/internal/auth"
	"github.com/KirillZiborov/GophKeeper/internal/config"
	"github.com/KirillZiborov/GophKeeper/internal/encryption"
	"github.com/KirillZiborov/GophKeeper/internal/models"
	"github.com/KirillZiborov/GophKeeper/internal/storage"
	"github.com/google/uuid"
)

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
	token, err := auth.GenerateToken(username)
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
		return "", errors.New("invalid username or password")
	}

	// Generates token for user.
	token, err := auth.GenerateToken(username)
	if err != nil {
		return "", err
	}

	return token, nil
}

// AddSecret adds secret data to user's list of credentials.
// Parameter token is a JWT which is used for etracting userID.
func (ks *KeeperService) AddSecret(ctx context.Context, token, data, meta string) error {
	// Extract userID from JWT using auth.GetUserID.
	userID := auth.GetUserID(token)
	if userID == "" {
		return errors.New("invalid token")
	}

	creds := &models.Secret{
		ID:     uuid.New().String(),
		Data:   data,
		Meta:   meta,
		UserID: userID,
	}

	return ks.Store.AddSecret(creds)
}

// EditSecret updates secret data using its id.
func (ks *KeeperService) EditSecret(ctx context.Context, token, id, data, meta string) error {
	userID := auth.GetUserID(token)
	if userID == "" {
		return errors.New("invalid token")
	}

	creds := &models.Secret{
		ID:     id,
		Data:   data,
		Meta:   meta,
		UserID: userID,
	}

	return ks.Store.EditSecret(creds)
}

// GetSecret retrieves all user's credentials.
// Parameter token is a JWT which is used for etracting userID.
func (ks *KeeperService) GetSecret(ctx context.Context, token string) ([]models.Secret, error) {
	// Extract userID from JWT using auth.GetUserID.
	userID := auth.GetUserID(token)
	if userID == "" {
		return nil, errors.New("invalid token")
	}

	// Retrieve all user's secrets from storage.
	creds, err := ks.Store.GetSecret(userID)
	if err != nil {
		return nil, err
	}
	return creds, nil
}
