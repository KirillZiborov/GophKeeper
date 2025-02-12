package models

// User provides a structure for registered user.
type User struct {
	ID       string `json:"id"`       // Unique user's id
	Username string `json:"username"` // Username
	Password string `json:"password"` // Hashed user's password
}

// Secret provides a structure for secret data.
type Secret struct {
	ID     int64  `json:"id"`      // Unique credentials id
	UserID string `json:"user_id"` // User's id
	Data   string `json:"data"`    // Secret data
	Meta   string `json:"meta"`    // Additional Metadata
}
