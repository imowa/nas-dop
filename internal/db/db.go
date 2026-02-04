package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB wraps a SQLite database connection.
type DB struct {
	conn *sql.DB
}

// Open opens a SQLite database at the given path with the specified busy timeout.
// It enables WAL mode and foreign keys.
func Open(dbPath string, busyTimeout time.Duration) (*DB, error) {
	timeoutMs := int(busyTimeout.Milliseconds())
	dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_timeout=%d", dbPath, timeoutMs)

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// Enable foreign keys
	if _, err := conn.Exec("PRAGMA foreign_keys = ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return &DB{conn: conn}, nil
}

// RunMigrations executes all SQL migration files.
func (db *DB) RunMigrations() error {
	data, err := migrationsFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("read migration file: %w", err)
	}

	if _, err := db.conn.Exec(string(data)); err != nil {
		return fmt.Errorf("execute migration: %w", err)
	}

	log.Println("migrations completed successfully")
	return nil
}

// DB returns the underlying *sql.DB for queries.
func (db *DB) DB() *sql.DB {
	return db.conn
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}
