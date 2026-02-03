package share

import (
	"crypto/rand"
	"encoding/base64"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Share represents a shared file or directory.
type Share struct {
	ID           int
	Token        string
	Path         string
	PasswordHash string
	ExpiresAt    *time.Time
	Name         string
	CreatedAt    time.Time
}

// GenerateToken generates a secure random token for a share (16 bytes, base64 URL-safe).
func GenerateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// ValidatePassword checks if the provided password matches the share's password hash.
func ValidatePassword(share *Share, password string) bool {
	if share.PasswordHash == "" {
		return true // No password required
	}
	err := bcrypt.CompareHashAndPassword([]byte(share.PasswordHash), []byte(password))
	return err == nil
}

// IsExpired checks if the share has expired.
func IsExpired(share *Share) bool {
	if share.ExpiresAt == nil {
		return false // No expiry
	}
	return time.Now().After(*share.ExpiresAt)
}
