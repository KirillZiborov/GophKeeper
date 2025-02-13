package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/KirillZiborov/GophKeeper/internal/logging"
	"github.com/KirillZiborov/GophKeeper/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when there is no data found.
var ErrNotFound = errors.New("not found")

// ErrAlreadyExists is returned when the data already exists.
var ErrAlreadyExists = errors.New("already exists")

// Storage defines interface for using PostgreSQL database.
type Storage interface {
	// Register a new user.
	RegisterUser(user *models.User) error
	// Authentificate user by his username and password. Returns User structure.
	GetUser(username string) (models.User, error)
	// Add new secret data for user with userID.
	AddSecret(cred *models.Secret) (int64, error)
	// Edit an existing secret data by his ID.
	EditSecret(creds *models.Secret) error
	// Returns a list of users secret data.
	GetSecret(userID string) ([]models.Secret, error)
}

// CreateURLTable initializes the 'users' table in the PostgreSQL database if it does not already exist
// as well as the 'secret'database.
// It defines the schema for storing users and their secrets.
//
// Parameters:
// - ctx: The context for managing cancellation and timeouts.
// - db: The PostgreSQL connection pool.
//
// Returns:
// - An error if the table creation fails; otherwise, nil.
func CreateTables(ctx context.Context, db *pgxpool.Pool) error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
    		uuid UUID PRIMARY KEY,
    		username TEXT NOT NULL UNIQUE,
    		password TEXT NOT NULL)`
	_, err := db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to create table: %w", err)
	}

	query = `
    CREATE TABLE IF NOT EXISTS secrets (
			id SERIAL PRIMARY KEY,
			user_id UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
			data TEXT NOT NULL,
			meta TEXT
		)`
	_, err = db.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("unable to create table: %w", err)
	}
	return nil
}

// DBStore represents a database store for URL records.
// It encapsulates the PostgreSQL connection pool to perform database operations.
type DBStore struct {
	db *pgxpool.Pool
}

// NewDBStore initializes and returns a pointer to a new instance of DBStore.
func NewDBStore(db *pgxpool.Pool) *DBStore {
	return &DBStore{db: db}
}

// RegisterUser adds a new user to the database.
func (store *DBStore) RegisterUser(user *models.User) error {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	err := store.db.QueryRow(context.Background(), query, user.Username).Scan(&exists)
	if err != nil {
		return err
	}
	if exists {
		return ErrAlreadyExists
	}

	query = `INSERT INTO users (uuid, username, password) VALUES ($1, $2, $3)`
	_, err = store.db.Exec(context.Background(), query, user.ID, user.Username, user.Password)

	if err != nil {
		return err
	}

	return nil
}

// GetUser retrieves users data by his username.
func (store *DBStore) GetUser(username string) (models.User, error) {
	var user models.User
	query := `SELECT * FROM users WHERE username=$1`
	err := store.db.QueryRow(context.Background(), query, username).Scan(&user.ID, &user.Username, &user.Password)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrNotFound
		}
		return models.User{}, err
	}

	return user, nil
}

// AddSecret saves users credentials to the database.
func (store *DBStore) AddSecret(cred *models.Secret) (int64, error) {
	query := `INSERT INTO secrets (user_id, data, meta) VALUES ($1, $2, $3) RETURNING id`
	var id int64
	err := store.db.QueryRow(context.Background(), query, cred.UserID, cred.Data, cred.Meta).Scan(&id)

	if err != nil {
		return 0, err
	}

	return id, nil
}

// EditSecret updates users credentials in the database.
func (store *DBStore) EditSecret(cred *models.Secret) error {
	query := `UPDATE secrets SET data = $1, meta = $2 WHERE id = $3`
	_, err := store.db.Exec(context.Background(), query, cred.Data, cred.Meta, cred.ID)

	if err != nil {
		return err
	}

	return nil
}

// GetSecret retrives and returns all users credentials.
func (store *DBStore) GetSecret(userID string) ([]models.Secret, error) {
	query := `SELECT id, user_id, data, meta FROM secrets WHERE user_id=$1`
	rows, err := store.db.Query(context.Background(), query, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	defer rows.Close()

	secret := make([]models.Secret, 0)
	for rows.Next() {
		var cred models.Secret
		err := rows.Scan(&cred.ID, &cred.UserID, &cred.Data, &cred.Meta)
		if err != nil {
			logging.Sugar.Errorw("failed to retrieve secret", "error", err)
			return nil, err
		}
		secret = append(secret, cred)
	}

	if err = rows.Err(); err != nil {
		logging.Sugar.Errorw("failed to retrieve secret", "error", err)
		return nil, err
	}

	return secret, nil
}
