package storage

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"
)

// GenerateThumbnail generates a thumbnail for an image file.
// Returns the thumbnail data or an error if the file is not an image.
func (s *Storage) GenerateThumbnail(relPath string, maxSize int) ([]byte, error) {
	// Resolve path
	absPath, err := s.resolvePath(relPath)
	if err != nil {
		return nil, err
	}

	// Check if it's an image by extension
	ext := strings.ToLower(filepath.Ext(absPath))
	if !isImageExt(ext) {
		return nil, fmt.Errorf("not an image file")
	}

	// Get file info for cache key
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	// Check cache
	cacheKey := generateCacheKey(relPath, info.ModTime(), maxSize)
	cacheDir := filepath.Join(s.root, ".thumbcache")
	cachePath := filepath.Join(cacheDir, cacheKey+".jpg")

	// If cache exists and is valid, return it
	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	// Generate thumbnail
	data, err := generateThumbnail(absPath, maxSize)
	if err != nil {
		return nil, err
	}

	// Save to cache
	os.MkdirAll(cacheDir, 0755)
	os.WriteFile(cachePath, data, 0644)

	return data, nil
}

// isImageExt checks if a file extension is an image type.
func isImageExt(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return true
	}
	return false
}

// generateCacheKey creates a cache key from path, mtime, and size.
func generateCacheKey(path string, modTime time.Time, maxSize int) string {
	h := md5.New()
	h.Write([]byte(path))
	h.Write([]byte(modTime.String()))
	h.Write([]byte(fmt.Sprintf("%d", maxSize)))
	return hex.EncodeToString(h.Sum(nil))
}

// generateThumbnail decodes, resizes, and encodes an image.
func generateThumbnail(srcPath string, maxSize int) ([]byte, error) {
	// Open source image
	file, err := os.Open(srcPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode image
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// Calculate new dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	var newWidth, newHeight int
	if width > height {
		newWidth = maxSize
		newHeight = height * maxSize / width
	} else {
		newHeight = maxSize
		newWidth = width * maxSize / height
	}

	// Resize image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Encode as JPEG
	tmpFile, err := os.CreateTemp("", "thumb-*.jpg")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if err := jpeg.Encode(tmpFile, dst, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}

	// Read back the encoded data
	tmpFile.Seek(0, 0)
	return os.ReadFile(tmpFile.Name())
}
