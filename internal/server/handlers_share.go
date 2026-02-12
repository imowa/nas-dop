package server

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"nas-dop/internal/share"
	"nas-dop/internal/storage"
)

// handleSharePage displays a shared file or directory.
func (s *Server) handleSharePage(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	// Load share
	sh, err := s.shareStore.GetByToken(token)
	if err != nil {
		log.Printf("share: failed to load share with token %q: %v", token, err)
		http.Error(w, "Share not found", 404)
		return
	}

	// Check if expired
	if share.IsExpired(sh) {
		http.Error(w, "Share has expired", 410)
		return
	}

	// Check password protection
	if sh.PasswordHash != "" {
		// Check if password validated in cookie
		cookie, err := r.Cookie("share_" + token)
		if err != nil || cookie.Value != "validated" {
			s.render(w, "share/share_password", map[string]interface{}{
				"Token": token,
				"Error": "",
			})
			return
		}
	}

	// List files
	files, err := s.storage.List(sh.Path)
	if err != nil {
		log.Printf("share: failed to list files for share %q at path %q: %v", token, sh.Path, err)
		http.Error(w, "Failed to list files", 500)
		return
	}

	s.render(w, "share/share", map[string]interface{}{
		"Token": token,
		"Name":  sh.Name,
		"Path":  sh.Path,
		"Files": files,
	})
}

// handleSharePassword validates share password.
func (s *Server) handleSharePassword(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	password := r.FormValue("password")

	// Load share
	sh, err := s.shareStore.GetByToken(token)
	if err != nil {
		log.Printf("share: failed to load share with token %q: %v", token, err)
		http.Error(w, "Share not found", 404)
		return
	}

	// Validate password
	if !share.ValidatePassword(sh, password) {
		s.render(w, "share/share_password", map[string]interface{}{
			"Token": token,
			"Error": "Invalid password",
		})
		return
	}

	// Set validation cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "share_" + token,
		Value:    "validated",
		Path:     "/share/" + token,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})

	http.Redirect(w, r, "/share/"+token, http.StatusSeeOther)
}

// handleShareDownload downloads a file from a share.
func (s *Server) handleShareDownload(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	filePath := r.PathValue("path")

	// Load and validate share
	sh, err := s.shareStore.GetByToken(token)
	if err != nil {
		http.Error(w, "Share not found", 404)
		return
	}

	// Check if expired
	if share.IsExpired(sh) {
		http.Error(w, "Share has expired", 410)
		return
	}

	// Check password protection
	if sh.PasswordHash != "" {
		cookie, err := r.Cookie("share_" + token)
		if err != nil || cookie.Value != "validated" {
			http.Error(w, "Unauthorized", 403)
			return
		}
	}

	// Verify path is within share path
	fullPath := filepath.Join(sh.Path, filePath)
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(sh.Path)) {
		http.Error(w, "Access denied", 403)
		return
	}

	// Read file
	data, err := s.storage.Read(fullPath)
	if err != nil {
		log.Printf("share: failed to read file %q for share %q: %v", fullPath, token, err)
		http.Error(w, "File not found", 404)
		return
	}

	filename := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

// handleShareThumb serves a thumbnail for an image file in a share.
func (s *Server) handleShareThumb(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	filePath := r.PathValue("path")

	// Load and validate share
	sh, err := s.shareStore.GetByToken(token)
	if err != nil {
		http.Error(w, "Share not found", 404)
		return
	}

	// Check if expired
	if share.IsExpired(sh) {
		http.Error(w, "Share has expired", 410)
		return
	}

	// Check password protection
	if sh.PasswordHash != "" {
		cookie, err := r.Cookie("share_" + token)
		if err != nil || cookie.Value != "validated" {
			http.Error(w, "Unauthorized", 403)
			return
		}
	}

	// Verify path is within share path
	fullPath := filepath.Join(sh.Path, filePath)
	if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(sh.Path)) {
		http.Error(w, "Access denied", 403)
		return
	}

	// Generate thumbnail
	data, err := s.storage.GenerateThumbnail(fullPath, s.cfg.ThumbMaxSizeShare)
	if err != nil {
		log.Printf("share: failed to generate thumbnail for %q in share %q: %v", fullPath, token, err)
		http.Error(w, "Thumbnail not available", 404)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

// handleShareZip creates a ZIP of selected files (Phase 3 - basic implementation).
func (s *Server) handleShareZip(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")

	// Load and validate share
	sh, err := s.shareStore.GetByToken(token)
	if err != nil {
		http.Error(w, "Share not found", 404)
		return
	}

	// Check if expired
	if share.IsExpired(sh) {
		http.Error(w, "Share has expired", 410)
		return
	}

	// Check password protection
	if sh.PasswordHash != "" {
		cookie, err := r.Cookie("share_" + token)
		if err != nil || cookie.Value != "validated" {
			http.Error(w, "Unauthorized", 403)
			return
		}
	}

	// Parse selected file paths from form
	if err := r.ParseForm(); err != nil {
		log.Printf("share: failed to parse form for ZIP in share %q: %v", token, err)
		http.Error(w, "Invalid form data", 400)
		return
	}

	paths := r.Form["paths[]"]
	if len(paths) == 0 {
		http.Error(w, "No files selected", 400)
		return
	}

	// Validate all paths are within share path
	var fullPaths []string
	for _, relPath := range paths {
		fullPath := filepath.Join(sh.Path, relPath)
		if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(sh.Path)) {
			http.Error(w, "Access denied", 403)
			return
		}
		fullPaths = append(fullPaths, fullPath)
	}

	// Create ZIP limits from config
	limits := storage.ZipLimits{
		MaxFiles: s.cfg.ZipMaxFiles,
		MaxBytes: s.cfg.ZipMaxBytes,
	}

	// Set headers
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sh.Name+".zip"))

	// Stream ZIP
	if err := s.storage.CreateZip(w, fullPaths, limits); err != nil {
		// Can't send error response after headers are sent
		// Log the error instead
		log.Printf("share: failed to create ZIP for share %q: %v", token, err)
		return
	}
}
