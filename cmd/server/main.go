// Studio photo-sharing app: single binary, SQLite, server-rendered HTML.
// See docs/plan-from-scratch.md and docs/build-roadmap.md.
package main

import (
	"database/sql"
	"log"
	"os"

	"nas-dop/internal/auth"
	"nas-dop/internal/config"
	"nas-dop/internal/db"
	"nas-dop/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := config.EnsureDirs(cfg); err != nil {
		log.Fatalf("ensure dirs: %v", err)
	}

	// Open database
	database, err := db.Open(cfg.DBPath, cfg.SQLiteBusyTimeout)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer database.Close()

	// Run migrations
	if err := database.RunMigrations(); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	// Bootstrap: create default admin if no users exist
	if err := bootstrapAdmin(database.DB(), cfg); err != nil {
		log.Fatalf("bootstrap: %v", err)
	}

	srv, err := server.New(cfg, database)
	if err != nil {
		log.Fatalf("server: %v", err)
	}

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := srv.Listen(addr); err != nil && err != os.ErrClosed {
		log.Fatalf("serve: %v", err)
	}
}

// bootstrapAdmin creates a default admin user if no users exist.
func bootstrapAdmin(db *sql.DB, cfg *config.Config) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Hash default password
		hash, err := auth.HashPassword(cfg.DefaultAdminPassword)
		if err != nil {
			return err
		}

		// Insert default admin
		_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", cfg.DefaultAdminUser, hash)
		if err != nil {
			return err
		}

		log.Printf("WARNING: Created default admin user '%s' - please change password immediately!", cfg.DefaultAdminUser)
	}

	return nil
}
