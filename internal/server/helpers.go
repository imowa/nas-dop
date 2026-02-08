package server

import (
	"fmt"
	"net/http"
	"strings"
)

// errRequestBodyTooLarge is the message Go's http.MaxBytesReader returns when limit is exceeded.
const errRequestBodyTooLarge = "http: request body too large"

// IsRequestEntityTooLarge reports whether err is from reading past MaxBytesReader limit.
// Handlers that read the body (e.g. upload) should use this and respond 413 when true.
func IsRequestEntityTooLarge(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), errRequestBodyTooLarge)
}

// WriteRequestEntityTooLarge sends 413 Request Entity Too Large with a clear message.
// Use when IsRequestEntityTooLarge(err) is true after reading the request body.
func WriteRequestEntityTooLarge(w http.ResponseWriter) {
	http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
}

// FormatBytes converts bytes to human-readable format (KB, MB, GB, etc.)
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
