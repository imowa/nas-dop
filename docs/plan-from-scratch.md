# Big Plan: Studio Photo-Sharing App from Scratch

Build a **studio photo-sharing** web app from scratch with a **modular project structure**. Same goals as the FileBrowser-based setup: one folder per customer, public share links (no login), **mobile-first UI** (most customers access share links on phone), GDrive-like and phone-friendly, download as ZIP or individual files. Targets **HG680FJ (Armbian + CasaOS)** and 2GB RAM.

---

## 1. Goals and scope

| Goal | Description |
|------|-------------|
| **Studio workflow** | One folder per customer; upload photos; share folder via link; customer views/downloads without account. |
| **Admin UI** | Log in, browse files/folders, upload/delete/rename, create folder, **create share** (link, optional expiry/password). |
| **Public share page** | Anyone with link can open `/share/<token>`, see folder contents; **normies-friendly download**: user can **select all** or **select multiple** photos, then **download selected** (each file, no unzip – primary) or **download as one file (ZIP)** (secondary, with "How to unzip?" tip). **Designed for mobile first** – most customers open the link on their phone. |
| **Look and feel** | **Mobile-first**: share page and customer flow optimized for small screens (single-column, large touch targets ~44–48px, readable text, no horizontal scroll). GDrive-like (clean layout, cards/grid, light theme). Desktop: responsive enhancements. |
| **Deploy** | Single Docker image (or backend + static frontends); ARM64/ARM32 + amd64; CasaOS-compatible compose; low memory. |

**Primary audience:** **Mobile.** Most customers access share links on their phone; the **share page UI must be mobile-first** (layout, touch targets, thumbnails, download buttons). Admin UI can be desktop-first; share page is the customer-facing experience.

**Explicit constraints**

- **Share links are read-only** – Customers can view and download only; no upload/delete via share link. Admin does all write operations after login.

**Out of scope for v1:** Desktop sync client, real-time collaboration, video transcoding, advanced permissions (e.g. per-user folders). Can add later as new modules.

---

## 2. Tech stack (lightweight for 2GB ARM)

| Layer | Recommendation | Why |
|-------|----------------|-----|
| **Backend** | **Go** (or Node.js) | Go: single binary, low RAM, easy ARM cross-compile. Node: same stack as frontend if you use JS. |
| **Database** | **SQLite** | No extra process; good for users + share tokens. |
| **Admin UI** | **Server-rendered HTML + HTMX** (or SPA with React/Vue) | HTMX: minimal JS, fast, works on slow devices. SPA: richer UX, more code. |
| **Share UI** | **Server-rendered HTML + minimal JS** | Same as admin: simple, fast, SEO-friendly (share links). |
| **Styling** | **CSS** (one codebase; **mobile-first** for share page, then desktop breakpoints) | Shared `static/css/`; GDrive-like; base = mobile, `min-width` for desktop. No heavy framework. |
| **File storage** | **Host filesystem** (e.g. `/data` or `/srv`) | One root dir; subdirs = folders; PUID/PGID for permissions. |
| **ZIP** | **Backend stream** (e.g. Go `archive/zip` or Node `archiver`) | Generate ZIP on the fly for “download folder as ZIP”. |

**Suggested default:** **Go** backend, **server-rendered HTML + HTMX** for both admin and share UIs, **SQLite**, **single binary + embedded static files** → one Docker image.

---

## 2b. Config and environment variables

| Variable | Purpose | Example |
|----------|---------|--------|
| `ROOT` | Filesystem root for files (all paths relative to this) | `/data` or `/srv` |
| `DB_PATH` | Path to SQLite database file | `/data/db/app.sqlite` |
| `SESSION_SECRET` | Secret for signing session cookies | long random string |
| `PORT` | HTTP listen port | `8080` or `80` |
| `DEFAULT_ADMIN_USER` | First-run admin username (only used when no users exist) | `admin` |
| `DEFAULT_ADMIN_PASSWORD` | First-run admin password (change immediately) | `admin` |
| `PUID`, `PGID` | (Docker) Owner for created files; match host user if needed | `1000`, `1000` |
| `APP_NAME` | Optional: app name in UI / branding | `Studio Photos` |

Load from env (and optional `.env` file). Run migrations at startup; if no users exist, create one from `DEFAULT_ADMIN_*` and redirect to change-password on first login.

---

## 3. Modular project structure

```
Nas-Dop/
├── README.md
├── go.mod                    # If Go backend (or package.json for Node)
├── .env.example
├── .gitignore
│
├── cmd/                      # Application entrypoints
│   └── server/
│       └── main.go           # Start backend (API + serve admin + share UIs)
│
├── internal/                 # Private application code (not importable)
│   ├── auth/                 # Auth module
│   │   ├── auth.go           # Login, session, password change
│   │   └── middleware.go     # RequireAuth, optional share-token
│   ├── storage/              # File storage module
│   │   ├── storage.go        # List, read, write, delete, mkdir; path safety
│   │   ├── zip.go            # Stream ZIP for folder download
│   │   └── thumb.go          # Generate/serve photo thumbnails for preview (resize, disk cache)
│   ├── share/                # Share links module
│   │   ├── share.go          # Create share, resolve token, list files for share
│   │   └── store.go          # SQLite: save/load share metadata (path, token, expiry, password)
│   ├── config/               # Config load (env, file)
│   │   └── config.go
│   └── server/               # HTTP server: routes, handlers
│       ├── server.go         # NewServer, Listen
│       ├── routes.go         # Route registration
│       ├── handlers_admin.go # Admin API + HTML handlers
│       ├── handlers_share.go # Public share handlers (HTML + download, ZIP)
│       └── handlers_api.go   # Optional JSON API for future SPA
│
├── pkg/                      # Optional: reusable packages (if needed)
│   └── ...
│
├── web/                      # Frontend assets and templates
│   ├── static/               # Static files
│   │   ├── css/
│   │   │   ├── admin.css     # Admin UI: GDrive-like + responsive
│   │   │   └── share.css     # Share page: mobile-first, same theme
│   │   ├── js/
│   │   │   ├── admin.js      # Optional: HTMX or minimal JS
│   │   │   └── share.js
│   │   └── img/              # Logo, favicons
│   ├── templates/            # Server-rendered HTML
│   │   ├── admin/            # Admin UI
│   │   │   ├── base.html
│   │   │   ├── login.html
│   │   │   ├── files.html   # File manager
│   │   │   └── share_create.html
│   │   └── share/            # Public share
│   │       ├── share.html    # List folder (with download ZIP / download each)
│   │       └── share_password.html
│   └── embed.go              # Go: embed static + templates (or pack at build)
│
├── migrations/               # DB migrations (SQLite)
│   └── 001_init.sql         # users, shares tables
│
├── docker/                   # Docker and deployment
│   ├── Dockerfile            # Multi-stage: build Go + static, run single binary
│   ├── docker-compose.yml    # Local dev
│   └── docker-compose.casaos.yml  # CasaOS: x-casaos metadata, volumes, port
│
├── config/                   # Config files (optional)
│   └── example.env
│
└── docs/                     # Documentation
    ├── plan-from-scratch.md  # This file
    ├── api.md                # API outline (for future SPA or integrations)
    ├── architecture.md       # Data flow, auth flow
    └── tutorial-install-casaos.md  # Existing tutorial (adapt for from-scratch app)
```

**Module responsibilities**

| Module | Responsibility |
|--------|----------------|
| **cmd/server** | Entrypoint: load config, init DB, storage, share store, start HTTP server. |
| **internal/auth** | Session-based login (cookie); password hash; middleware to protect admin routes. |
| **internal/storage** | All file operations under one root; path traversal safety; optional ZIP stream; **thumbnail** generate + disk cache (`thumb.go`); optional chown to PUID/PGID on write (for Docker). |
| **internal/share** | Generate token; store (path, token, expiry, password hash); resolve token → path + options. |
| **internal/server** | Register routes; admin handlers (login, file list, upload, delete, create share); share handlers (public page, download file, download ZIP). |
| **web/** | HTML templates + CSS + JS; **mobile-first** for share page (customer-facing); GDrive-like; two “apps”: admin and share. |
| **docker/** | Build and run for local dev and CasaOS (ARM + amd64). |

---

## 4. Architecture (high level)

```mermaid
flowchart TB
  subgraph Client
    Browser[Browser]
  end
  subgraph App
    Router[HTTP Router]
    Admin[Admin Handlers]
    Share[Share Handlers]
    Auth[Auth Middleware]
    Storage[Storage Module]
    ShareStore[Share Store]
    DB[(SQLite)]
  end
  subgraph Host
    Root[/data or /srv]
  end
  Browser -->|"/" admin| Router
  Browser -->|"/share/TOKEN" public| Router
  Router --> Auth
  Auth -->|admin routes| Admin
  Auth -->|share routes no auth| Share
  Admin --> Storage
  Admin --> ShareStore
  Share --> ShareStore
  Share --> Storage
  ShareStore --> DB
  Storage --> Root
```

- **Admin:** `/` → login; `/files`, `/files/*` → file manager; `/share/new` → create share (form: path, optional expiry, password). Auth required.
- **Share:** `/share/<token>` → public page (list folder; **checkboxes** for select all / select multiple; **“Download selected”** (primary, each file) and **“Download as one file (ZIP)”** with “How to unzip?” tip). `/share/<token>/dl/<path>` → download single file. `/share/<token>/zip?paths=...` → stream ZIP of **selected** files only. No login; optional password gate.

---

## 5. Data model

**SQLite**

- **users** – `id`, `username`, `password_hash`, `created_at`. Single admin for v1 (or a few).
- **shares** – `id`, `token` (unique), `path` (relative to root), `password_hash` (nullable), `expires_at` (nullable), `created_at`. Optional: `name` (label for admin).

**Filesystem**

- One **root** (e.g. `/data` or env `ROOT`). All listed paths are relative to root; no path traversal (validate, no `..`).

---

## 6. API / routes outline

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/` | No | Redirect to login or /files |
| GET | `/login` | No | Login form |
| POST | `/login` | No | Authenticate, set session |
| POST | `/logout` | Yes | Clear session |
| GET | `/files` | Yes | File manager (list root) |
| GET | `/files/*path` | Yes | List folder or download file (if ?dl=1) |
| POST | `/files/*path` | Yes | Upload file(s), create folder |
| DELETE | `/files/*path` | Yes | Delete file/folder |
| PATCH | `/files/*path` | Yes | Rename (body: new name) |
| GET | `/share/new` | Yes | Form: pick folder, expiry, password |
| POST | `/share/new` | Yes | Create share; redirect to “share link” page |
| GET | `/share/<token>` | No | Public share page (list folder); optional password form |
| POST | `/share/<token>` | No | Submit password if required |
| GET | `/share/<token>/dl/*path` | No | Download single file (path relative to share root) |
| GET | `/share/<token>/zip` | No | Download **selected** files as one ZIP (query `?paths=file1,file2,...` or POST body). If no selection, optional: whole folder. Stream; do not buffer full ZIP in RAM. |
| GET | `/files/thumb/*path` | Yes | Thumbnail for image (resized, e.g. max 320px); for admin file list preview |
| GET | `/share/<token>/thumb/*path` | No | Thumbnail for image in share; for share page grid preview |
| GET | `/health` or `/ready` | No | Health check for Docker/CasaOS (return 200 if app is up) |

Optional later: JSON API under `/api/` for SPA or mobile (same logic, JSON responses).

---

## 6a. Normies-friendly download (select all / select multiple, one file or each)

Make download **normies-friendly**: user can **select all** or **select multiple** photos, then get **each file separately** (no unzip) or **one file** (ZIP). **Many people don't know how to unzip**, so we make **"Download selected"** (each file, no ZIP) the **primary** option and ZIP secondary, with a **"How to unzip?"** tip when they choose ZIP.

### Recommendation

| Option | What | Normies-friendly? |
|--------|------|-------------------|
| **A. Download each (selected)** | Checkboxes; “Select all” / “Select none” / pick multiple. One button: **“Download selected”** – each selected file downloads separately (no ZIP). | **Yes.** No unzip needed; photos go straight to device. Best for people who don’t know how to unzip. |
| **B. Select → Download as ZIP (selected)** | Same checkboxes; button **“Download as one file (ZIP)”**. Server returns one ZIP containing only the selected files. | **Yes** for “one file”; **add “How to unzip?”** help because some users don’t know how. |
| **C. Remove ZIP entirely** | Only “Download each”. | Optional; keeping ZIP gives “one file” for those who prefer it. |

**Recommended flow:** **“Download selected” (each file, no ZIP) as primary.** Add checkboxes; “Select all”; **“Download selected”** (primary) – saves each photo separately, no unzip. Second button: **“Download as one file (ZIP)”** (secondary) – one ZIP; show a **“How to unzip?”** tip/link next to it (phone: tap file → Open → Save to Photos or Files; computer: right‑click ZIP → Extract).

### Implementation

| Item | What to do |
|------|------------|
| **Share page UI** | Each file/photo has a **checkbox**. Toolbar: **“Select all”**, **“Select none”**, **“Download selected”** (primary – each file, no unzip), **“Download as one file (ZIP)”** (secondary). Next to ZIP button: short **“How to unzip?”** tip or collapsible help. |
| **ZIP of selection** | `GET /share/<token>/zip?paths=file1.jpg,file2.jpg,...` or `POST /share/<token>/zip` with body `{"paths":["file1.jpg","file2.jpg"]}`. Paths relative to share root; validate each. Stream ZIP containing only those files. If no paths (or empty), optional: whole folder (backward compatible) or return 400 “select at least one file”. |
| **Download each** | “Download each” opens a page listing download links for each selected file, or runs JS to trigger `window.location` / `<a download>` for each (with small delay to reduce pop-up blocking). Document that “Download as ZIP” is recommended for many files. |
| **Copy** | Primary: “Select the photos you want, then click **Download selected** to save each photo separately (no unzip).” Secondary: “Or click **Download as one file (ZIP)** to get a single file; [How to unzip?](#).” |

### Summary

- **Primary:** Select all or multiple → **“Download selected”** → each file downloads separately; no unzip – best for people who don't know how to unzip.
- **Secondary:** **“Download as one file (ZIP)”** – one file; show **“How to unzip?”** tip next to the button or in help (phone: tap file → Open → Save to Photos; computer: right‑click ZIP → Extract).
- ZIP stays as an option for users who want one file and are okay unzipping (or after reading the tip).

---

## 6b. Photo thumbnails for preview

When listing files (admin file manager or public share page), show **photo thumbnails** so users can preview images without opening the full file. Non-image files use an icon or placeholder.

### Approach

| Step | What to do |
|------|------------|
| **1. Thumbnail endpoint** | **Admin:** `GET /files/thumb/*path` (auth required). **Share:** `GET /share/<token>/thumb/*path`. Path validated same as download (under root / share root). |
| **2. Only for images** | If file is not an image (e.g. by extension or MIME: jpeg, png, webp, gif), return 404 or redirect to a generic “file” icon so the UI can show a fallback. |
| **3. Generate or serve cached** | **First request:** Decode image, resize to a max dimension (e.g. 320px on longest side), encode as JPEG (quality ~85), write to **disk cache** (e.g. `ROOT/.thumb/<hash-of-path>.jpg` or a dedicated cache dir). Stream to response. **Later requests:** If cache file exists and source file mtime unchanged, serve cache file (stream). Otherwise regenerate. |
| **4. Headers** | **Content-Type:** `image/jpeg`. **Cache-Control:** e.g. `private, max-age=86400` (1 day) so browsers cache. **Content-Length:** set if serving from file. |
| **5. Memory** | Do **not** load full-res image into memory; decode, resize in a bounded buffer (e.g. decode to RGBA, resize to max 320px, encode). Use Go `image` package (or `resize`, `imaging`) with a max dimension to cap memory on 2GB ARM. |

### Cache invalidation

- Cache key: path + file modification time (mtime). If the file is updated (re-upload), mtime changes and thumbnail is regenerated on next request.
- Optional: delete cache entry when file is deleted (or lazy: 404 on next thumb request).

### UI usage

- **Admin file list / share grid:** For each item with image extension, use `<img src="/files/thumb/path/to/photo.jpg" alt="">` or `/share/<token>/thumb/path/to/photo.jpg`. CSS: max-width/height (e.g. 200px), object-fit cover. Fallback: show icon or placeholder if thumb returns 404.
- **Lazy loading:** Use `loading="lazy"` on `<img>` so thumbnails load as they enter viewport (saves bandwidth on long lists).

### Summary

| Item | Choice |
|------|--------|
| **Endpoint** | `/files/thumb/*path` (admin), `/share/<token>/thumb/*path` (share). |
| **Format** | Resized JPEG, max 320px (or 200px) on longest side. |
| **Cache** | Disk cache (path + mtime); serve cached file when valid. |
| **Non-images** | Return 404; UI shows icon. |
| **Memory** | Bounded resize; no full-res in memory. |

---

## 6c. Download handling (how user file downloads work)

When a user (admin or customer) downloads a file, the server does the following so downloads are safe, predictable, and work well on phones and desktops.

### Single-file download (admin and share)

| Step | What to do |
|------|------------|
| **1. Resolve path** | **Admin:** path = request path under `ROOT`; **Share:** path = share root + request path. Reject if path contains `..` or leaves root/share root. |
| **2. Check it’s a file** | Stat path; if not a regular file (e.g. directory, symlink), return 404 or 400. |
| **3. Open and stream** | Open file for reading; **stream** chunk-by-chunk to the response (e.g. `io.Copy(responseWriter, file)` in Go). Do **not** read the whole file into memory (keeps RAM low and supports large files). |
| **4. Set headers** | **Content-Disposition:** `attachment; filename="<safe-name>"` – use **basename only** (no path) for the filename to avoid path injection. **Content-Type:** detect from extension (e.g. `image/jpeg`, `application/pdf`) or use `application/octet-stream`. **Content-Length:** set if known (file size) so browsers show progress. |
| **5. Cleanup** | Close file handle when done (defer in Go). |

**Admin:** e.g. `GET /files/photos/image.jpg?dl=1` → resolve `photos/image.jpg` under ROOT, stream file, `Content-Disposition: attachment; filename="image.jpg"`.

**Share:** e.g. `GET /share/<token>/dl/subfolder/image.jpg` → resolve share root + `subfolder/image.jpg`, validate path under share root, stream file, same headers.

### Download selected as ZIP (share only) – normies-friendly

- Handler: `GET /share/<token>/zip?paths=file1.jpg,file2.jpg,...` or `POST /share/<token>/zip` with body `{"paths":["file1.jpg","file2.jpg"]}`. Paths are relative to share root; validate each (no `..`).
- **Stream** the ZIP to the response: write only the selected files into the ZIP; do **not** build the full ZIP in memory.
- **Content-Disposition:** `attachment; filename="selected.zip"` or `"<folder-name>-selected.zip"`.
- **Content-Type:** `application/zip`.
- If no paths (or empty): either return 400 “Select at least one file” or, for backward compatibility, ZIP whole folder. Prefer “select at least one” for clarity.
- **UI:** Share page has checkboxes; “Select all”; **“Download selected”** (primary) and **“Download as one file (ZIP)”** (secondary) with **“How to unzip?”** tip.

### “Download each” (no ZIP) – primary for normies

- Same checkboxes; button **“Download selected”**: show a list of download links for each selected file (`GET /share/<token>/dl/<path>`), or use JS to trigger each link (with delay to reduce pop-up blocking). **Primary option** – no unzip; good for people who don’t know how to unzip.
- Wording: “Save each photo separately (no unzip).”

### Summary

| Scenario | Handler | Path validation | Response |
|----------|---------|-----------------|----------|
| Admin downloads file | GET /files/*path?dl=1 | path under ROOT | Stream file; Content-Disposition (basename); Content-Type; Content-Length |
| Share: single file | GET /share/<token>/dl/*path | path under share root | Same as admin |
| Share: selected as ZIP | GET /share/<token>/zip?paths=... or POST body | paths under share root | Stream ZIP of selected files only; Content-Disposition (selected.zip); application/zip |
| Share: “each file” | Multiple GET /share/<token>/dl/*path | per file | Same as single file |

---

## 7. Phases and milestones

### Phase 1: Backend core (2–3 weeks)

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

### Phase 2: Admin file manager (1–2 weeks)

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

### Phase 3: Public share page and download options (1–2 weeks)

- [ ] **Share page UI**: cards/list for files and folders; **checkboxes** on each file (Select all / Select none / pick multiple); **photo thumbnails** via `GET /share/<token>/thumb/*path` for images (resized, disk-cached); “Download” per file; **“Download selected”** (primary – each file, no unzip); **“Download as one file (ZIP)”** (secondary) with **“How to unzip?”** tip.
- [ ] **internal/storage/thumb.go**: Generate thumbnail (decode image, resize max 320px, encode JPEG); disk cache keyed by path + mtime; serve from cache when valid. Only for image MIME types (jpeg, png, webp, gif).
- [ ] **ZIP**: handler `GET /share/<token>/zip?paths=...` (or POST) – **stream** ZIP of **selected files only** (paths in query or body; validate under share root). If no selection, return 400 or optional whole folder. Do not buffer full ZIP in memory.
- [ ] **Share subfolders**: Share page supports browsing into subfolders (e.g. `?path=subfolder` or `/share/<token>/browse/path`); download and ZIP respect current path under share root.
- [ ] Share password: if set, show password form; POST → verify, set cookie/session for token, then show content.
- [ ] Expiry: on load share, check `expires_at`; if past, show “link expired”.
- [ ] **Mobile-first share page**: Touch targets ~44–48px; single-column; no horizontal scroll; test on real phone (primary) and narrow viewport. Desktop = responsive enhancement.

**Milestone:** Customer opens link (optional password), sees folder with checkboxes; can **select all** or **select multiple** → **“Download selected”** (each file, no unzip – primary) or **“Download as one file (ZIP)”** (secondary, with “How to unzip?” tip). Single-file download still available per item.

### Phase 4: Polish and CasaOS (1 week)

- [ ] **Styling**: Final pass – **mobile-first** share page (layout, tap targets, thumbnails); GDrive-like colors, typography, spacing; shared CSS; desktop breakpoints.
- [ ] **Branding**: Logo, favicon (web/static/img); config or env for app name.
- [ ] **Security**: Rate limit (e.g. login, share resolve); secure cookies (HttpOnly, SameSite); CSRF tokens on forms (login, share create, share password); no path traversal anywhere; share download path validated under share root.
- [ ] **Health endpoint**: `GET /health` or `GET /ready` returns 200 when app is up (for Docker healthcheck and CasaOS).
- [ ] **Docker**: Multi-stage build; ARM64/ARM32 + amd64; single binary + embedded static/templates.
- [ ] **docker-compose.casaos.yml**: CasaOS `x-casaos` metadata; volumes `/DATA/AppData/$AppID/db`, `/DATA` (or custom) for root; port 10180; env PUID, PGID, TZ; optional healthcheck using `/health`.
- [ ] **Docs**: Update tutorial (install on CasaOS, make public); short API doc (routes table).

**Milestone:** App runs in Docker on CasaOS; public share links work from LAN (and internet if reverse proxy used).

### Phase 5 (optional): Extras

- [ ] JSON API for admin (e.g. `/api/files`, `/api/shares`) for future SPA or scripts.
- [ ] Multiple users (admin list, per-user home dir).
- [ ] Share management in admin: list shares, revoke (delete token).

---

## 8. Security checklist

- [ ] **Path traversal**: Reject any path containing `..` or absolute; resolve under root only. For share downloads, validate path is under share root.
- [ ] **Auth**: Admin routes require session; share routes only need valid token (and password if set).
- [ ] **Passwords**: Bcrypt (or Argon2) for user and share password.
- [ ] **Session**: HttpOnly, Secure (if HTTPS), SameSite cookie; rotate on login.
- [ ] **CSRF**: Use CSRF tokens on all state-changing forms (login, create share, share password).
- [ ] **Token**: Long random token (e.g. 32 bytes hex); no guessable IDs.
- [ ] **Rate limit**: Login and share-resolve endpoints (e.g. 10/min per IP).
- [ ] **Upload limits**: Max file size and max request body size (e.g. 100MB / 500MB); return 413 when exceeded.
- [ ] **File types**: Optional: block executable MIME types on upload; serve downloads with safe `Content-Disposition` (filename only, no path).
- [ ] **Thumbnails**: Only generate for allowed image types (jpeg, png, webp, gif); bounded resize (max 320px) to limit memory; disk cache to avoid repeated work.

---

## 9. CasaOS integration

- **Compose:** Use `docker-compose.casaos.yml` with CasaOS `x-casaos` (architectures: amd64, arm64, arm; main service; title, description, icon, port_map; envs for PUID, PGID, TZ; volumes for db and root).
- **Volumes:** Map `/DATA/AppData/$AppID/db` → app’s SQLite dir; `/DATA` (or `/mnt/usb/share`) → app’s file root.
- **First run:** If no DB, run migrations then create default admin from env (`DEFAULT_ADMIN_USER`, `DEFAULT_ADMIN_PASSWORD`); redirect to change-password on first login.
- **Health:** Expose `GET /health` for Docker/CasaOS healthcheck.
- **Branding:** Logo/favicon in `web/static/img`; app name in config or env.

---

## 10. Summary

| Item | Content |
|------|---------|
| **Structure** | Modular: `cmd/`, `internal/` (auth, storage, share, server), `web/` (templates, static), `docker/`, `migrations/`. |
| **Stack** | Go (or Node), SQLite, server-rendered HTML + HTMX (or minimal JS), CSS. |
| **Phases** | 1 Backend core → 2 Admin file manager → 3 Share page + ZIP + download options → 4 Polish + CasaOS → 5 Optional extras. |
| **Deliverables** | Single binary (or Node app) + static/templates; Docker image (multi-arch); CasaOS compose; docs and tutorial. |

This plan is the “big plan” for building the app from scratch with a clear, modular layout. You can implement phase by phase and adjust tech choices (e.g. Node instead of Go, or SPA for admin) while keeping the same structure and goals.

---

## 11. Risks and mitigations

| Risk | Mitigation |
|------|------------|
| **ZIP uses too much RAM** | Stream ZIP to response; never load full archive in memory. Go `archive/zip` supports writing to `io.Writer`. |
| **Large uploads / DoS** | Enforce max file size and max request size; return 413. Consider timeouts. |
| **Share token leaked** | Tokens are unguessable; optional expiry and password limit impact. Document “revoke” in Phase 5 (delete share). |
| **Path traversal** | Validate every path: no `..`, no absolute; resolve under root (admin) or under share root (share). |
| **First-run default password** | Use env for default; force change-password on first login; document in tutorial. |
| **Thumbnail memory** | Resize with a max dimension (e.g. 320px); do not decode huge images into full RGBA in one go if avoidable; use disk cache so repeat views don’t regenerate. |
