package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"nas-dop/internal/auth"
	"nas-dop/internal/config"
	"nas-dop/internal/db"
	"nas-dop/internal/share"
	"nas-dop/internal/storage"
	"nas-dop/web"
)

// Server holds HTTP server and dependencies (auth, storage, share store, etc.).
type Server struct {
	cfg          *config.Config
	mux          *http.ServeMux
	db           *db.DB
	sessionStore *auth.SessionStore
	storage      *storage.Storage
	shareStore   *share.Store
	templates    *template.Template
}

// New builds a new Server with all dependencies wired.
func New(cfg *config.Config, database *db.DB) (*Server, error) {
	// Parse templates from embedded FS
	tmpl, err := template.ParseFS(web.FS, "templates/**/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	s := &Server{
		cfg:          cfg,
		mux:          http.NewServeMux(),
		db:           database,
		sessionStore: auth.NewSessionStore(),
		storage:      storage.New(cfg.Root, cfg.PUID, cfg.PGID),
		shareStore:   share.NewStore(database.DB()),
		templates:    tmpl,
	}
	s.routes()
	return s, nil
}

// Listen starts the HTTP server on addr (e.g. ":8080").
// Uses ReadHeaderTimeout, ReadTimeout, WriteTimeout from config (optimization-recommendations.md).
func (s *Server) Listen(addr string) error {
	handler := s.requestLimitMiddleware(s.mux)
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: s.cfg.ReadHeaderTimeout,
		ReadTimeout:       s.cfg.ReadTimeout,
		WriteTimeout:      s.cfg.WriteTimeout,
	}
	return srv.ListenAndServe()
}

// requestLimitMiddleware limits request body size for POST/PUT/PATCH (return 413 when exceeded).
func (s *Server) requestLimitMiddleware(next http.Handler) http.Handler {
	max := s.cfg.MaxRequestBytes
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			r.Body = http.MaxBytesReader(w, r.Body, max)
		}
		next.ServeHTTP(w, r)
	})
}

// render executes a template with the given data.
func (s *Server) render(w http.ResponseWriter, name string, data interface{}) {
	if err := s.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
		http.Error(w, "Internal Server Error", 500)
	}
}

// sessionDuration is the default session duration (24 hours).
const sessionDuration = 24 * time.Hour
