package share

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Store manages share persistence in SQLite.
type Store struct {
	db *sql.DB
}

// NewStore creates a new share store.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create creates a new share and returns it with the generated token.
func (s *Store) Create(path, name, password string, expiresAt *time.Time) (*Share, error) {
	token := GenerateToken()

	// Hash password if provided
	var passwordHash string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		passwordHash = string(hash)
	}

	// Insert into database
	result, err := s.db.Exec(
		"INSERT INTO shares (token, path, password_hash, expires_at, name) VALUES (?, ?, ?, ?, ?)",
		token, path, passwordHash, expiresAt, name,
	)
	if err != nil {
		return nil, fmt.Errorf("insert share: %w", err)
	}

	id, _ := result.LastInsertId()

	return &Share{
		ID:           int(id),
		Token:        token,
		Path:         path,
		PasswordHash: passwordHash,
		ExpiresAt:    expiresAt,
		Name:         name,
		CreatedAt:    time.Now(),
	}, nil
}

// GetByToken retrieves a share by its token.
func (s *Store) GetByToken(token string) (*Share, error) {
	var share Share
	var expiresAt sql.NullTime

	err := s.db.QueryRow(
		"SELECT id, token, path, password_hash, expires_at, name, created_at FROM shares WHERE token = ?",
		token,
	).Scan(&share.ID, &share.Token, &share.Path, &share.PasswordHash, &expiresAt, &share.Name, &share.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("share not found")
	}
	if err != nil {
		return nil, fmt.Errorf("query share: %w", err)
	}

	if expiresAt.Valid {
		share.ExpiresAt = &expiresAt.Time
	}

	return &share, nil
}

// Delete removes a share by its token.
func (s *Store) Delete(token string) error {
	_, err := s.db.Exec("DELETE FROM shares WHERE token = ?", token)
	return err
}

// List returns all shares (for admin UI).
func (s *Store) List() ([]*Share, error) {
	rows, err := s.db.Query(
		"SELECT id, token, path, password_hash, expires_at, name, created_at FROM shares ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shares []*Share
	for rows.Next() {
		var share Share
		var expiresAt sql.NullTime

		if err := rows.Scan(&share.ID, &share.Token, &share.Path, &share.PasswordHash, &expiresAt, &share.Name, &share.CreatedAt); err != nil {
			continue
		}

		if expiresAt.Valid {
			share.ExpiresAt = &expiresAt.Time
		}

		shares = append(shares, &share)
	}

	return shares, nil
}
