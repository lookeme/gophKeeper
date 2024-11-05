// Package models defines the data structures used for managing URL shortening and user details.
package models

import (
	"github.com/golang-jwt/jwt/v5"
	"time"
)

// UserDTO represents the data transfer object for a User entity.
type UserDTO struct {
	ID        string    `json:"id"`       // UUID of the user
	Username  string    `json:"username"` // Username of the user
	Email     string    `json:"email"`    // Email of the user
	Password  string    `json:"password"`
	Role      string    `json:"role"`       // Role of the user (e.g., admin, user)
	CreatedAt time.Time `json:"created_at"` // Timestamp of when the user was created
	UpdatedAt time.Time `json:"updated_at"` // Timestamp of when the user was last updated
}

// FileDTO represents the data transfer object for a File entity.
type FileDTO struct {
	ID        string    `json:"id"`         // UUID of the file
	FileName  string    `json:"file_name"`  // Name of the file
	FilePath  string    `json:"file_path"`  // Path or URL where the file is stored
	FileSize  int64     `json:"file_size"`  // Size of the file in bytes
	FileType  string    `json:"file_type"`  // MIME type of the file
	OwnerID   string    `json:"owner_id"`   // UUID of the owner (User)
	CreatedAt time.Time `json:"created_at"` // Timestamp of when the file was created
	UpdatedAt time.Time `json:"updated_at"` // Timestamp of when the file was last updated
}

type SettingsDTO struct {
	Key   string `json:"key"`   // UUID of the file
	Value string `json:"value"` // Name of the file
}

// Claims represents the custom claims for JWT authentication.
type Claims struct {
	UserID string
	jwt.RegisteredClaims
}

type CredData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type CreditCardData struct {
	CardNumber string `json:"card_number"`
	CardHolder string `json:"card_holder"`
	CardExp    string `json:"card_exp"`
	CardCVV    string `json:"card_cvv"`
	CardType   string `json:"card_type"`
}

type UserCredentials struct {
	Name      string    `json:"name"`
	Data      string    `json:"data"`
	DataType  string    `json:"type"`
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
}
