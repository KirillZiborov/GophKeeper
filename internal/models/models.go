// Package models provides internal structures representing users and their secrets.
package models

// User represents a registered user.
type User struct {
	ID       string `json:"id"`       // Unique user's id
	Username string `json:"username"` // Username
	Password string `json:"password"` // Hashed user's password
}

// Secret represents secret data.
type Secret struct {
	ID     int64  `json:"id"`      // Unique credentials id
	UserID string `json:"user_id"` // User's id
	Data   string `json:"data"`    // Secret data
	Meta   string `json:"meta"`    // Additional Metadata
}
