package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// Storage provides safe file operations within a root directory.
type Storage struct {
	root string
	puid int
	pgid int
}

// FileInfo represents file metadata.
type FileInfo struct {
	Name    string
	Size    int64
	ModTime time.Time
	IsDir   bool
	Ext     string // File extension (e.g., ".jpg", ".pdf")
}

// New creates a new Storage instance.
func New(root string, puid, pgid int) *Storage {
	return &Storage{
		root: filepath.Clean(root),
		puid: puid,
		pgid: pgid,
	}
}

// resolvePath validates and resolves a path to an absolute path within root.
// Accepts both relative paths and paths with leading slashes (treated as relative to root).
func (s *Storage) resolvePath(relPath string) (string, error) {
	// Clean and normalize the path
	cleaned := filepath.Clean(relPath)
	
	// Remove leading slash if present (treat as relative to root)
	// This allows paths like "/customer" and "customer" to work identically
	cleaned = strings.TrimPrefix(cleaned, "/")
	cleaned = strings.TrimPrefix(cleaned, "\\") // Windows support
	
	// Handle empty path (root)
	if cleaned == "." || cleaned == "" {
		cleaned = ""
	}
	
	// Check for path traversal attempts
	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}
	
	// Join with root
	absPath := filepath.Join(s.root, cleaned)
	absPath = filepath.Clean(absPath)
	
	// Verify result is within root (prevent traversal)
	rootClean := filepath.Clean(s.root)
	if !strings.HasPrefix(absPath, rootClean) {
		return "", fmt.Errorf("path outside root directory")
	}
	
	return absPath, nil
}

// List returns the contents of a directory.
func (s *Storage) List(relPath string) ([]FileInfo, error) {
	absPath, err := s.resolvePath(relPath)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Extract file extension (lowercase)
		ext := strings.ToLower(filepath.Ext(entry.Name()))

		files = append(files, FileInfo{
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsDir:   entry.IsDir(),
			Ext:     ext,
		})
	}

	return files, nil
}

// Read reads a file and returns its contents.
func (s *Storage) Read(relPath string) ([]byte, error) {
	absPath, err := s.resolvePath(relPath)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(absPath)
}

// Write writes data to a file, creating parent directories if needed.
func (s *Storage) Write(relPath string, data []byte) error {
	absPath, err := s.resolvePath(relPath)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Write file
	if err := os.WriteFile(absPath, data, 0644); err != nil {
		return err
	}

	// Apply PUID/PGID if configured (Docker use case)
	if s.puid > 0 || s.pgid > 0 {
		if err := os.Chown(absPath, s.puid, s.pgid); err != nil {
			// Log but don't fail (may not have permissions)
		}
	}

	return nil
}

// Delete removes a file or directory.
func (s *Storage) Delete(relPath string) error {
	absPath, err := s.resolvePath(relPath)
	if err != nil {
		return err
	}

	return os.RemoveAll(absPath)
}

// Mkdir creates a directory.
func (s *Storage) Mkdir(relPath string) error {
	absPath, err := s.resolvePath(relPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return err
	}

	// Apply PUID/PGID if configured
	if s.puid > 0 || s.pgid > 0 {
		if err := os.Chown(absPath, s.puid, s.pgid); err != nil {
			// Log but don't fail
		}
	}

	return nil
}

// Exists checks if a path exists.
func (s *Storage) Exists(relPath string) bool {
	absPath, err := s.resolvePath(relPath)
	if err != nil {
		return false
	}

	_, err = os.Stat(absPath)
	return err == nil
}

// WriteMultipart writes uploaded file data to storage.
func (s *Storage) WriteMultipart(relPath string, data []byte) error {
	return s.Write(relPath, data)
}

// Stat returns file info for a path.
func (s *Storage) Stat(relPath string) (*FileInfo, error) {
	absPath, err := s.resolvePath(relPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	}, nil
}

// Rename renames a file or directory.
func (s *Storage) Rename(oldPath, newName string) error {
	// Validate old path
	oldAbsPath, err := s.resolvePath(oldPath)
	if err != nil {
		return err
	}

	// Validate new name (should not contain path separators)
	if strings.Contains(newName, "/") || strings.Contains(newName, "\\") {
		return fmt.Errorf("new name cannot contain path separators")
	}

	// Construct new path (same directory, new name)
	dir := filepath.Dir(oldAbsPath)
	newAbsPath := filepath.Join(dir, newName)

	// Rename
	return os.Rename(oldAbsPath, newAbsPath)
}

// chown is a helper to apply ownership (Windows-safe).
func chown(path string, uid, gid int) error {
	if uid <= 0 && gid <= 0 {
		return nil
	}
	return syscall.Chown(path, uid, gid)
}
