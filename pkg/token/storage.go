// Package token provides functionalities for storing user's authentification tokens.
package token

import (
	"fmt"
	"io"
	"os"
)

// Storage is the interaface for saving and loading a token from file.
type Storage interface {
	// Load загружает токен
	Load() (accessToken string, err error)
	// Save saves token.
	Save(accessToken string) error
}

// FileStorage is a file storage for token.
type FileStorage struct {
	Path string
}

var _ Storage = (*FileStorage)(nil)

// NewFileStorage creates a new file for storing token.
func NewFileStorage(path string) *FileStorage {
	return &FileStorage{
		Path: path,
	}
}

// Save wrires a token to file.
func (s *FileStorage) Save(accessToken string) error {
	file, err := os.Create(s.Path)
	if err != nil {
		return err
	}

	defer func() {
		if err = file.Close(); err != nil {
			fmt.Println("Failed to close file storing token", err)
		}
	}()

	_, err = file.WriteString(accessToken)
	return err
}

// Load reads a token from file.
func (s *FileStorage) Load() (string, error) {
	file, err := os.Open(s.Path)
	if err != nil {
		return "", nil
	}
	defer func() {
		if err = file.Close(); err != nil {
			fmt.Println("Failed to close file storing token", err)
		}
	}()

	b, err := io.ReadAll(file)
	return string(b), err
}
