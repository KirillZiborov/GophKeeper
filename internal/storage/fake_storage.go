package storage

import (
	"errors"
	"sync"

	"github.com/KirillZiborov/GophKeeper/internal/models"
)

// FakeStorage implements Storage interface for tests using in-memory cards.
type FakeStorage struct {
	mu           sync.Mutex
	usersByName  map[string]*models.User             // username key
	usersByID    map[string]*models.User             // ID key
	secrets      map[string]map[int64]*models.Secret // UserID key, map ID -> Secret values
	nextSecretID int64
}

// NewFakeStorage creates a new instance of FakeStorage.
func NewFakeStorage() *FakeStorage {
	return &FakeStorage{
		usersByName:  make(map[string]*models.User),
		usersByID:    make(map[string]*models.User),
		secrets:      make(map[string]map[int64]*models.Secret),
		nextSecretID: 1,
	}
}

// RegisterUser adds a new user to the storage.
func (fs *FakeStorage) RegisterUser(user *models.User) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.usersByName[user.Username]; exists {
		return errors.New("user already exists")
	}
	fs.usersByName[user.Username] = user
	fs.usersByID[user.ID] = user
	return nil
}

// GetUser retrieves users data by his username.
func (fs *FakeStorage) GetUser(username string) (models.User, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	user, exists := fs.usersByName[username]
	if !exists {
		return models.User{}, ErrNotFound
	}
	return *user, nil
}

// AddSecret saves users credentials to the database.
func (fs *FakeStorage) AddSecret(secret *models.Secret) (int64, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.usersByID[secret.UserID]; !exists {
		return 0, errors.New("user not found")
	}

	secret.ID = fs.nextSecretID
	fs.nextSecretID++

	if fs.secrets[secret.UserID] == nil {
		fs.secrets[secret.UserID] = make(map[int64]*models.Secret)
	}
	fs.secrets[secret.UserID][secret.ID] = secret
	return secret.ID, nil
}

// EditSecret updates users credentials in the storage.
func (fs *FakeStorage) EditSecret(secret *models.Secret) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	userSecrets, exists := fs.secrets[secret.UserID]
	if !exists {
		return errors.New("no secrets for user")
	}
	if _, exists := userSecrets[secret.ID]; !exists {
		return errors.New("secret not found")
	}
	userSecrets[secret.ID] = secret
	return nil
}

// GetSecret retrives and returns all users credentials.
func (fs *FakeStorage) GetSecret(userID string) ([]models.Secret, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	userSecrets, exists := fs.secrets[userID]
	if !exists {
		return []models.Secret{}, nil
	}
	var secrets []models.Secret
	for _, secret := range userSecrets {
		secrets = append(secrets, *secret)
	}
	return secrets, nil
}
