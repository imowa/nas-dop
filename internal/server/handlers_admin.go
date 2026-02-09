package server

import (
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"nas-dop/internal/auth"
)

// handleLoginForm renders the login page.
func (s *Server) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	s.render(w, "admin/login", map[string]interface{}{
		"Error": "",
	})
}

// handleLoginPost processes login form submission.
func (s *Server) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Query user from database
	var passwordHash string
	err := s.db.DB().QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&passwordHash)
	if err == sql.ErrNoRows {
		s.render(w, "admin/login", map[string]interface{}{
			"Error": "Invalid username or password",
		})
		return
	}
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	// Validate password
	if !auth.CheckPassword(passwordHash, password) {
		s.render(w, "admin/login", map[string]interface{}{
			"Error": "Invalid username or password",
		})
		return
	}

	// Create session
	session, err := s.sessionStore.Create(username, sessionDuration)
	if err != nil {
		http.Error(w, "Internal server error", 500)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    session.Token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  session.ExpiresAt,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// handleLogout logs out the user.
func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		s.sessionStore.Delete(cookie.Value)
	}

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// handleAdminRoot redirects to /files.
func (s *Server) handleAdminRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/files", http.StatusSeeOther)
}

// handleFilesList lists files in a directory.
func (s *Server) handleFilesList(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")

	files, err := s.storage.List(path)
	if err != nil {
		http.Error(w, "Failed to list files", 500)
		return
	}

	// Build breadcrumbs
	breadcrumbs := buildBreadcrumbs(path)

	s.render(w, "admin/files", map[string]interface{}{
		"Path":        path,
		"Files":       files,
		"Breadcrumbs": breadcrumbs,
	})
}

// handleUpload handles file uploads.
func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")

	// Parse multipart form
	err := r.ParseMultipartForm(s.cfg.MaxUploadBytes)
	if err != nil {
		if IsRequestEntityTooLarge(err) {
			WriteRequestEntityTooLarge(w)
			return
		}
		http.Error(w, "Failed to parse form", 400)
		return
	}

	files := r.MultipartForm.File["files"]
	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}

		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			continue
		}

		// Write file to storage
		filePath := filepath.Join(path, fileHeader.Filename)
		if err := s.storage.Write(filePath, data); err != nil {
			continue
		}
	}

	http.Redirect(w, r, "/files/"+strings.TrimPrefix(path, "/"), http.StatusSeeOther)
}

// handleMkdir creates a new directory.
func (s *Server) handleMkdir(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	name := r.FormValue("name")

	if name == "" {
		http.Error(w, "Name required", 400)
		return
	}

	newPath := filepath.Join(path, name)
	if err := s.storage.Mkdir(newPath); err != nil {
		http.Error(w, "Failed to create directory", 500)
		return
	}

	http.Redirect(w, r, "/files/"+strings.TrimPrefix(path, "/"), http.StatusSeeOther)
}

// handleDelete deletes a file or directory.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")

	if err := s.storage.Delete(path); err != nil {
		http.Error(w, "Failed to delete", 500)
		return
	}

	// Redirect to parent directory
	parent := filepath.Dir(path)
	http.Redirect(w, r, "/files/"+strings.TrimPrefix(parent, "/"), http.StatusSeeOther)
}

// handleRename renames a file or directory.
func (s *Server) handleRename(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	newName := r.FormValue("newName")

	if newName == "" {
		http.Error(w, "New name required", 400)
		return
	}

	if err := s.storage.Rename(path, newName); err != nil {
		http.Error(w, "Failed to rename", 500)
		return
	}

	// Redirect to parent directory
	parent := filepath.Dir(path)
	http.Redirect(w, r, "/files/"+strings.TrimPrefix(parent, "/"), http.StatusSeeOther)
}

// handleDownload serves a file for download.
func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")

	data, err := s.storage.Read(path)
	if err != nil {
		http.Error(w, "File not found", 404)
		return
	}

	filename := filepath.Base(path)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(data)
}

// handleFilesThumb serves a thumbnail for an image file.
func (s *Server) handleFilesThumb(w http.ResponseWriter, r *http.Request) {
	path := r.PathValue("path")

	data, err := s.storage.GenerateThumbnail(path, s.cfg.ThumbMaxSizeAdmin)
	if err != nil {
		http.Error(w, "Thumbnail not available", 404)
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

// handleShareForm renders the share creation form.
func (s *Server) handleShareForm(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")

	s.render(w, "admin/share_create", map[string]interface{}{
		"Path":    path,
		"Success": "",
		"ShareURL": "",
	})
}

// handleShareCreate creates a new share.
func (s *Server) handleShareCreate(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	name := r.FormValue("name")
	password := r.FormValue("password")
	expiresStr := r.FormValue("expires")

	var expiresAt *time.Time
	if expiresStr != "" {
		t, err := time.Parse("2006-01-02", expiresStr)
		if err == nil {
			expiresAt = &t
		}
	}

	share, err := s.shareStore.Create(path, name, password, expiresAt)
	if err != nil {
		http.Error(w, "Failed to create share", 500)
		return
	}

	// Detect protocol from request
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	shareURL := fmt.Sprintf("%s://%s/share/%s", scheme, r.Host, share.Token)

	s.render(w, "admin/share_create", map[string]interface{}{
		"Path":     path,
		"Success":  "Share created successfully!",
		"ShareURL": shareURL,
	})
}

// handleSharesList displays all shares for management.
func (s *Server) handleSharesList(w http.ResponseWriter, r *http.Request) {
	shares, err := s.shareStore.List()
	if err != nil {
		http.Error(w, "Failed to list shares", 500)
		return
	}

	// Detect protocol for base URL
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)

	s.render(w, "admin/shares", map[string]interface{}{
		"Shares":  shares,
		"BaseURL": baseURL,
		"Success": "",
	})
}

// handleShareDelete deletes a share.
func (s *Server) handleShareDelete(w http.ResponseWriter, r *http.Request) {
	token := r.FormValue("token")

	if err := s.shareStore.Delete(token); err != nil {
		http.Error(w, "Failed to delete share", 500)
		return
	}

	http.Redirect(w, r, "/shares?success=deleted", http.StatusSeeOther)
}

// buildBreadcrumbs creates breadcrumb navigation from a path.
func buildBreadcrumbs(path string) []map[string]string {
	if path == "/" || path == "" {
		return []map[string]string{
			{"Name": "Home", "Path": "/"},
		}
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")
	breadcrumbs := []map[string]string{
		{"Name": "Home", "Path": "/"},
	}

	currentPath := ""
	for _, part := range parts {
		if part == "" {
			continue
		}
		currentPath = filepath.Join(currentPath, part)
		breadcrumbs = append(breadcrumbs, map[string]string{
			"Name": part,
			"Path": "/" + currentPath,
		})
	}

	return breadcrumbs
}
