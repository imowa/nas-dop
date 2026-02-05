package storage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
)

// ZipLimits holds limits for ZIP creation.
type ZipLimits struct {
	MaxFiles int   // Maximum number of files
	MaxBytes int64 // Maximum total bytes
}

// CreateZip streams a ZIP archive to w containing the specified files.
// Paths are relative to the storage root and are validated.
// Returns an error if limits are exceeded or if any file cannot be read.
func (s *Storage) CreateZip(w io.Writer, paths []string, limits ZipLimits) error {
	// Check file count limit
	if len(paths) > limits.MaxFiles {
		return fmt.Errorf("too many files: %d exceeds limit of %d", len(paths), limits.MaxFiles)
	}

	// Create ZIP writer
	zw := zip.NewWriter(w)
	defer zw.Close()

	var totalBytes int64

	for _, relPath := range paths {
		// Resolve and validate path
		absPath, err := s.resolvePath(relPath)
		if err != nil {
			return fmt.Errorf("invalid path %s: %w", relPath, err)
		}

		// Get file info
		info, err := os.Stat(absPath)
		if err != nil {
			return fmt.Errorf("stat %s: %w", relPath, err)
		}

		// Skip directories (could be enhanced to recursively add directory contents)
		if info.IsDir() {
			continue
		}

		// Check size limit
		if totalBytes+info.Size() > limits.MaxBytes {
			return fmt.Errorf("total size exceeds limit of %d bytes", limits.MaxBytes)
		}

		// Add file to ZIP
		if err := s.addFileToZip(zw, absPath, relPath); err != nil {
			return fmt.Errorf("add %s to zip: %w", relPath, err)
		}

		totalBytes += info.Size()
	}

	return nil
}

// addFileToZip adds a single file to the ZIP archive.
func (s *Storage) addFileToZip(zw *zip.Writer, absPath, relPath string) error {
	// Open source file
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create ZIP entry
	writer, err := zw.Create(relPath)
	if err != nil {
		return err
	}

	// Copy file contents to ZIP
	_, err = io.Copy(writer, file)
	return err
}
