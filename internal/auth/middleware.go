package auth

import (
	"context"
	"net/http"
)

type contextKey string

const usernameKey contextKey = "username"

// RequireAuth is middleware that requires a valid session cookie.
// If the session is invalid, it redirects to /login.
// If valid, it adds the username to the request context.
func RequireAuth(store *SessionStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read session cookie
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate session
		session, ok := store.Get(cookie.Value)
		if !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add username to context
		ctx := context.WithValue(r.Context(), usernameKey, session.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUsername extracts the username from the request context.
func GetUsername(r *http.Request) string {
	username, _ := r.Context().Value(usernameKey).(string)
	return username
}
