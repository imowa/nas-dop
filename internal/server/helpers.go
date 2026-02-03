package server

import (
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
