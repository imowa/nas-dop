# Implementation status

What’s implemented vs still TODO. Track phases in **[build-roadmap.md](build-roadmap.md)**.

---

## Implemented

| Area | What | Where |
|------|------|--------|
| **Config** | All env vars (ROOT, DB_PATH, session, port, admin defaults, PUID/PGID, **optimization**) | `internal/config/config.go` |
| **HTTP server** | Listen with **ReadHeaderTimeout**, **ReadTimeout**, **WriteTimeout** | `internal/server/server.go` |
| **Request body limit** | Middleware limits POST/PUT/PATCH to MaxRequestBytes; **413 helpers** for handlers | `server.go`, `helpers.go` |
| **Routes** | `GET /health` → 200 OK; `GET /static/*` → embedded static with **Cache-Control** | `routes.go`, `handlers_health.go`, `handlers_static.go` |
| **Migrations** | SQL file for `users`, `shares` (not run by app yet) | `migrations/001_init.sql` |
| **Docker** | Dockerfile (multi-stage), docker-compose, docker-compose.casaos.yml | `docker/` |
| **Web** | Embedded static + templates (templates not rendered yet) | `web/embed.go`, `web/static/`, `web/templates/` |
| **Ensure dirs** | Create ROOT and DB directory parent at startup so storage/DB don’t fail with "dir not found" | `config.EnsureDirs`, called from `main.go` |

---

## Not implemented (TODO)

### Phase 1 – Backend core

| Item | Where | Notes |
|------|--------|------|
| Open SQLite, run migrations | `cmd/server/main.go` or new `internal/db` | Use `cfg.DBPath`, `cfg.SQLiteBusyTimeout` |
| ~~Ensure ROOT and DB dir exist at startup~~ | ~~main.go or config~~ | **Done:** `config.EnsureDirs(cfg)` in main |
| **internal/storage** | List dir, read/write file, delete, mkdir; path validation (no `..`) | `internal/storage/storage.go` |
| **internal/auth** | Bcrypt, session cookie, login/logout, change-password | `internal/auth/auth.go` |
| **internal/auth** | RequireAuth middleware for admin routes | `internal/auth/middleware.go` |
| **internal/share** | Create token, save/load share (SQLite), resolve token → path, expiry, password | `share.go`, `store.go` |
| **internal/server** | Wire auth, storage, share store, DB into Server | `server.go` New() |
| **Routes** | `/`, `/login`, `/logout`, `/files`, `/files/*`, `/share/new` | `routes.go` |
| **Handlers** | Login form/POST, file list (root), create share form/POST | `handlers_admin.go` |
| **Handlers** | Share page (resolve token, list path); password form | `handlers_share.go` |
| **Templates** | Parse/execute embedded templates (login, files, share) | Need template loader + render in handlers |
| **Bootstrap** | If no users → create default admin from env, redirect to change-password on first login | After DB + auth |

### Phase 2 – Admin file manager

| Item | Where |
|------|--------|
| Files UI (breadcrumbs, navigate, download, thumbnails) | handlers_admin, templates |
| Upload, mkdir, delete, rename | storage + handlers_admin |
| Create share form with copy-link | handlers_admin, share store |
| Enforce `MaxUploadBytes` per file; use 413 helpers | handlers_admin |

### Phase 3 – Share page + download

| Item | Where |
|------|--------|
| Share page UI (checkboxes, Download selected, Download as ZIP, “How to unzip?”) | handlers_share, templates |
| **internal/storage/thumb.go** | Generate + disk cache; use ThumbMaxSizeShare/Admin, ThumbConcurrency |
| **internal/storage/zip.go** | Stream ZIP of selected files; enforce ZipMaxFiles, ZipMaxBytes |
| Share subfolders, password, expiry | handlers_share, share store |

### Phase 4 – Polish

| Item | Where |
|------|--------|
| Styling, branding, security (CSRF, rate limit), docs | Multiple |

---

## Quick checklist

- [x] Config + optimization env
- [x] HTTP timeouts + request body limit + 413 helpers
- [x] GET /health, GET /static/* with Cache-Control
- [x] Migrations SQL file
- [x] Docker + CasaOS compose
- [x] Ensure ROOT and DB dir at startup
- [ ] DB open + migrations run
- [ ] Auth (login, session, middleware)
- [ ] Storage (list, read, write, delete, mkdir)
- [ ] Share store + token create/resolve
- [ ] Admin routes + handlers + templates
- [ ] Share routes + handlers + templates
- [ ] Thumb + ZIP handlers
