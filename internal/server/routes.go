package server

import (
	"log"
	"net/http"

	"nas-dop/internal/auth"
)

// routes registers all HTTP routes. See docs/plan-from-scratch.md §6.
func (s *Server) routes() {
	// Health (no auth)
	s.mux.HandleFunc("GET /health", s.handleHealth)

	// Static assets with Cache-Control (optimization-recommendations.md)
	if h, err := s.staticHandler(); err == nil {
		s.mux.Handle("GET /static/", h)
	} else {
		log.Printf("static handler: %v (skipping /static/)", err)
	}

	// Login/logout (no auth required)
	s.mux.HandleFunc("GET /login", s.handleLoginForm)
	s.mux.HandleFunc("POST /login", s.handleLoginPost)
	s.mux.HandleFunc("POST /logout", s.handleLogout)

	// Admin routes (require auth)
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /", s.handleAdminRoot)
	adminMux.HandleFunc("GET /files", s.handleFilesList)
	adminMux.HandleFunc("GET /files/{path...}", s.handleFilesList)
	adminMux.HandleFunc("POST /files/upload", s.handleUpload)
	adminMux.HandleFunc("POST /files/mkdir", s.handleMkdir)
	adminMux.HandleFunc("POST /files/delete", s.handleDelete)
	adminMux.HandleFunc("POST /files/rename", s.handleRename)
	adminMux.HandleFunc("GET /files/download/{path...}", s.handleDownload)
	adminMux.HandleFunc("GET /shares", s.handleSharesList)
	adminMux.HandleFunc("POST /shares/delete", s.handleShareDelete)
	adminMux.HandleFunc("GET /files/thumb/{path...}", s.handleFilesThumb)

	s.mux.Handle("/", auth.RequireAuth(s.sessionStore, adminMux))

	// Share creation routes (require auth) - must be registered before wildcard /share/{token}
	s.mux.Handle("GET /share/new", auth.RequireAuth(s.sessionStore, http.HandlerFunc(s.handleShareForm)))
	s.mux.Handle("POST /share/new", auth.RequireAuth(s.sessionStore, http.HandlerFunc(s.handleShareCreate)))

	// Share routes (public, no auth)
	s.mux.HandleFunc("GET /share/{token}", s.handleSharePage)
	s.mux.HandleFunc("POST /share/{token}/password", s.handleSharePassword)
	s.mux.HandleFunc("GET /share/{token}/dl/{path...}", s.handleShareDownload)
	s.mux.HandleFunc("POST /share/{token}/zip", s.handleShareZip)
	s.mux.HandleFunc("GET /share/{token}/thumb/{path...}", s.handleShareThumb)
}
