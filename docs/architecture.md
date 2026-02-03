# Architecture

High-level data flow and auth flow. Details in **[plan-from-scratch.md](plan-from-scratch.md)** §4–5.

## Data flow

- **Admin:** Browser → Router → Auth (session) → Admin handlers → Storage / Share store → SQLite or filesystem (ROOT).
- **Share:** Browser → Router → Share handlers (token ± password) → Share store (resolve path) → Storage (read/list) → response.
- **Storage:** All file operations under one ROOT; path validation (no `..`). Thumbnails: generate + disk cache; ZIP: stream to response.

## Auth flow

- **Admin:** Login (POST /login) → verify password (bcrypt) → set session cookie → redirect to /files. Logout clears cookie. Middleware protects /files, /share/new.
- **Share:** No login. Optional share password: GET /share/<token> shows password form; POST verifies, sets cookie/session for that token, then shows content. Expiry: if `expires_at` is past, show “link expired”.

## Data model

- **SQLite:** `users` (id, username, password_hash, created_at), `shares` (id, token, path, password_hash, expires_at, name, created_at).
- **Filesystem:** One root (ROOT); paths relative to root; no path traversal.
