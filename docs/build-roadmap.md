# Build roadmap: Studio photo-sharing app from scratch

Track progress against **[plan-from-scratch.md](plan-from-scratch.md)**. Check off items as they’re done.

---

## Phase 1: Backend core (2–3 weeks)

- [ ] Project layout: `cmd/`, `internal/`, `web/` (placeholder templates).
- [ ] Config: root path, DB path, session secret, port (env).
- [ ] SQLite: migrations for `users`, `shares`; open DB at startup.
- [ ] **internal/storage**: list dir, read file, write file, delete, mkdir; strict path validation (no `..`).
- [ ] **internal/auth**: password hash (bcrypt), session (cookie), login/logout, middleware.
- [ ] **internal/share**: create token (secure random), save/load share; resolve token → path + expiry/password check.
- [ ] **internal/server**: routes; admin handlers (login, list root only first); share handler (resolve token, list path). No ZIP yet.
- [ ] **Bootstrap**: At startup, run migrations; if no users exist, create default admin from env (`DEFAULT_ADMIN_USER`, `DEFAULT_ADMIN_PASSWORD`) and redirect to change-password on first login.
- [ ] **Error handling**: Consistent error responses: 404 (not found), 403 (forbidden), 500 (generic); “Share not found” or “Link expired” for invalid/expired token.
- [ ] Minimal **templates**: login form, one “files” page (list root), one “share” page (list share path), simple error page.
- [ ] **Docker**: Dockerfile (Go build + run); docker-compose for local dev. Run and test on host.

**Milestone:** Log in, see root folder, create share, open share link and see folder (no download yet).

---

## Phase 2: Admin file manager (1–2 weeks)

- [ ] **Files UI**: list any path (breadcrumbs), click folder = navigate; click file = download (or preview link). **Thumbnails:** for image files, show thumbnail via `GET /files/thumb/*path`; use `<img src="..." loading="lazy">`; fallback to icon for non-images.
- [ ] Upload: form or drag-drop (multipart); save to current path.
- [ ] Create folder: form (name) → mkdir.
- [ ] Delete: button + confirm; delete file or recursive delete folder.
- [ ] Rename: inline or form.
- [ ] **Create share**: form on “share new” – pick path (current or type), optional expiry, optional password; POST → create share, show link with **copy-to-clipboard** button.
- [ ] **Upload limits**: Max file size and max request size (e.g. 100MB per file, 500MB per request) to avoid DoS; return 413 with clear message.
- [ ] **PUID/PGID**: When creating files/dirs (upload, mkdir), chown to PUID/PGID if set (so host user owns files in Docker).
- [ ] **CSS**: **Mobile-first** for share page (base = small screen; `min-width` for desktop). GDrive-like layout; admin can be desktop-first. Touch targets ~44–48px on share page.

**Milestone:** Full admin file manager + create share with link; share page lists files; download single file from share page.

---

## Phase 3: Public share page and download options (1–2 weeks)

- [ ] **Share page UI**: cards/list for files and folders; **checkboxes** on each file (Select all / Select none / pick multiple); **photo thumbnails** via `GET /share/<token>/thumb/*path` for images (resized, disk-cached); “Download” per file; **“Download selected”** (primary – each file, no unzip); **“Download as one file (ZIP)”** (secondary) with **“How to unzip?”** tip.
- [ ] **internal/storage/thumb.go**: Generate thumbnail (decode image, resize max 320px, encode JPEG); disk cache keyed by path + mtime; serve from cache when valid. Only for image MIME types (jpeg, png, webp, gif).
- [ ] **ZIP**: handler `GET /share/<token>/zip?paths=...` (or POST) – **stream** ZIP of **selected files only** (paths in query or body; validate under share root). If no selection, return 400 or optional whole folder. Do not buffer full ZIP in memory.
- [ ] **Share subfolders**: Share page supports browsing into subfolders (e.g. `?path=subfolder` or `/share/<token>/browse/path`); download and ZIP respect current path under share root.
- [ ] Share password: if set, show password form; POST → verify, set cookie/session for token, then show content.
- [ ] Expiry: on load share, check `expires_at`; if past, show “link expired”.
- [ ] **Mobile-first share page**: Touch targets ~44–48px; single-column; no horizontal scroll; test on real phone (primary) and narrow viewport. Desktop = responsive enhancement.

**Milestone:** Customer opens link (optional password), sees folder with checkboxes; can **select all** or **select multiple** → **“Download selected”** (each file, no unzip – primary) or **“Download as one file (ZIP)”** (secondary, with “How to unzip?” tip). Single-file download still available per item.

---

## Phase 4: Polish and CasaOS (1 week)

- [ ] **Styling**: Final pass – **mobile-first** share page (layout, tap targets, thumbnails); GDrive-like colors, typography, spacing; shared CSS; desktop breakpoints.
- [ ] **Branding**: Logo, favicon (web/static/img); config or env for app name.
- [ ] **Security**: Rate limit (e.g. login, share resolve); secure cookies (HttpOnly, SameSite); CSRF tokens on forms (login, share create, share password); no path traversal anywhere; share download path validated under share root.
- [ ] **Health endpoint**: `GET /health` or `GET /ready` returns 200 when app is up (for Docker healthcheck and CasaOS).
- [ ] **Docker**: Multi-stage build; ARM64/ARM32 + amd64; single binary + embedded static/templates.
- [ ] **docker-compose.casaos.yml**: CasaOS `x-casaos` metadata; volumes `/DATA/AppData/$AppID/db`, `/DATA` (or custom) for root; port 10180; env PUID, PGID, TZ; optional healthcheck using `/health`.
- [ ] **Docs**: Update tutorial (install on CasaOS, make public); short API doc (routes table).

**Milestone:** App runs in Docker on CasaOS; public share links work from LAN (and internet if reverse proxy used).

---

## Phase 5 (optional): Extras

- [ ] JSON API for admin (e.g. `/api/files`, `/api/shares`) for future SPA or scripts.
- [ ] Multiple users (admin list, per-user home dir).
- [ ] Share management in admin: list shares, revoke (delete token).
