# Optimization recommendations for file sharing

Suggestions to make the studio photo-sharing app faster, lighter, and better on mobile and low-RAM (2GB ARM). The main plan is in **[plan-from-scratch.md](plan-from-scratch.md)**; this doc adds **optional** optimizations.

---

## 1. Downloads and streaming

| Recommendation | Why |
|----------------|-----|
| **Always stream** | Use `io.Copy` (or equivalent) for single-file and ZIP; never read whole file(s) into RAM. Plan already says this; enforce it in code review. |
| **Set Content-Length** | For file and cached-thumbnail responses, set `Content-Length` so browsers show progress and don’t buffer unnecessarily. |
| **Range requests (optional)** | For large files, support `Range` so users can resume and video can seek. Add later if needed; not required for photos. |
| **Download “each file” in batches** | If JS triggers many `/dl/<path>` in a row, add a small delay (e.g. 200–500 ms) between opens to avoid connection limits and pop-up blocking; or show a page of direct links instead of auto-downloading. |

---

## 2. Thumbnails

| Recommendation | Why |
|----------------|-----|
| **Disk cache + mtime** | Cache key = path + source mtime (plan §6b). Serve cached file with `Content-Length` and `Cache-Control: private, max-age=86400`. |
| **Smaller thumbs on share page** | Plan says max 320px; for share grid on mobile, 200px is often enough and saves CPU + bandwidth. Consider 200px for share, 320px for admin. |
| **Lazy load** | Use `loading="lazy"` on `<img>` (plan already says this). Reduces initial load on long lists. |
| **Limit concurrent thumb generation** | On 2GB ARM, cap how many thumbnails are generated at once (e.g. semaphore or worker pool) so one heavy folder doesn’t OOM. |
| **Optional: WebP thumbs** | Serve thumbnails as WebP where supported (Accept header); smaller than JPEG. Add only if you want extra bandwidth savings. |

---

## 3. ZIP

| Recommendation | Why |
|----------------|-----|
| **Stream only** | Use `archive/zip` writing to `http.ResponseWriter`; never build the full ZIP in memory (plan §11). |
| **Limit selection size (optional)** | Cap number of files or total size per ZIP request (e.g. 500 files or 2 GB) to avoid long-running requests and timeouts; return 400 with a clear message. |
| **Flush periodically** | For large ZIPs, flush the response buffer periodically so the client gets data earlier and doesn’t time out. |

---

## 4. Mobile share page

| Recommendation | Why |
|----------------|-----|
| **Touch targets 44–48px** | Plan says this; keep buttons and checkboxes at least 44px so taps register reliably. |
| **Pagination or virtual list** | For shares with hundreds of files, show a limited list per page or “load more” so the first HTML and thumb requests stay bounded. |
| **Prefer “Download selected”** | Primary action = each file, no ZIP (plan §6a). Makes the main path simple and avoids unzip confusion. |
| **Sticky toolbar** | Keep “Select all”, “Download selected”, “Download as ZIP” in a fixed bar so they’re always reachable without scrolling. |

---

## 5. Server and memory (2GB ARM)

| Recommendation | Why |
|----------------|-----|
| **Read timeouts** | Set `ReadHeaderTimeout` and `ReadTimeout` on the HTTP server so slow or stuck clients don’t hold connections forever. |
| **Upload limits** | Enforce max file size and max body size (plan §8); return 413. E.g. 100MB per file, 500MB per request. |
| **SQLite busy timeout** | Use a small busy timeout (e.g. 5s) so concurrent requests don’t fail immediately under load. |
| **Thumb cache location** | Put thumb cache on the same volume as ROOT (or a dedicated cache dir) so it’s on disk, not in RAM. |

---

## 6. Caching and headers

| Recommendation | Why |
|----------------|-----|
| **Static assets** | Serve CSS/JS with long `Cache-Control` (e.g. `max-age=86400`) and use a versioned path or query (e.g. `admin.css?v=1`) when you change them. |
| **Share page HTML** | Short cache or no-store for the share page itself (it’s dynamic). Thumb and download URLs can have longer cache. |
| **304 Not Modified** | For file download and thumb endpoints, support `If-None-Match` / `If-Modified-Since` and return 304 when unchanged; reduces bandwidth. |

---

## 7. Docker image

| Recommendation | Why |
|----------------|-----|
| **Multi-stage build** | Already in plan; keep build deps out of the final image. |
| **Small base** | Alpine is fine; avoid adding unnecessary packages. Only add what’s needed (e.g. `wget` for healthcheck, `ca-certificates`). |
| **Single binary** | Embed static and templates in the binary (plan); no extra COPY of files in production. |

---

## 8. Optional later

- **CDN / cache in front** | If share links are public, put a reverse proxy (Caddy, Nginx) with caching in front; cache thumbnails and static assets, not the share HTML or download endpoints.
- **Compression** | Enable gzip (or Brotli) for HTML, CSS, JS; do not compress already-compressed (JPEG, PNG, ZIP) responses.
- **Share listing cache** | For a share’s file list, short in-memory cache (e.g. 30s) if you expect many refreshes; invalidate on any change. Only add if you measure a need.

---

## Summary

| Area | Priority | Action |
|------|----------|--------|
| Streaming + Content-Length | High | Implement from day one for files and ZIP. |
| Thumb disk cache + lazy load | High | Per plan; add concurrency limit for thumbs on ARM. |
| ZIP stream + optional cap | High | Stream only; consider file/size limit. |
| Mobile touch targets + sticky toolbar | High | Per plan; improves share UX. |
| Timeouts + upload limits | High | Protects server on 2GB. |
| 304 / cache headers | Medium | Add when handlers are in place. |
| Range / WebP / pagination | Low | Add only if you need them. |

These are **recommendations**; the plan remains the source of truth. Implement what fits your timeline and measure on real devices (especially phone and HG680FJ) when you can.

---

## Implemented in the app

The following are already wired so handlers can use them:

| Item | Where | Env / config |
|------|--------|-------------|
| **HTTP timeouts** | `internal/server/server.go` | `ReadHeaderTimeout` (10s), `ReadTimeout` (30s), `WriteTimeout` (60s) via `READ_HEADER_TIMEOUT`, `READ_TIMEOUT`, `WRITE_TIMEOUT` |
| **Request body limit** | `internal/server/server.go`, `helpers.go` | Middleware limits POST/PUT/PATCH body to `MaxRequestBytes` (500MB). Handlers that read the body should call `server.IsRequestEntityTooLarge(err)` and `server.WriteRequestEntityTooLarge(w)` to send 413. Env: `MAX_REQUEST_BYTES` |
| **Upload / ZIP / thumb limits** | `internal/config/config.go` | `MaxUploadBytes` (100MB), `ThumbMaxSizeShare` (200), `ThumbMaxSizeAdmin` (320), `ThumbConcurrency` (4), `ZipMaxFiles` (500), `ZipMaxBytes` (2GB), `SQLiteBusyTimeout` (5s). Env: `MAX_UPLOAD_BYTES`, `THUMB_*`, `ZIP_MAX_*`, `SQLITE_BUSY_TIMEOUT` |
| **Static Cache-Control** | `internal/server/handlers_static.go`, `routes.go` | `GET /static/*` served with `Cache-Control: private, max-age=<STATIC_CACHE_MAX_AGE>` (default 86400). Env: `STATIC_CACHE_MAX_AGE` |

Upload handler (Phase 2) should enforce `cfg.MaxUploadBytes` per file and use `IsRequestEntityTooLarge` / `WriteRequestEntityTooLarge` when parsing the multipart body; thumb handler should use `cfg.ThumbConcurrency` (semaphore) and `ThumbMaxSizeShare` / `ThumbMaxSizeAdmin`; ZIP handler should enforce `cfg.ZipMaxFiles` and `cfg.ZipMaxBytes`. DB open (Phase 1) should set `_busy_timeout` to `cfg.SQLiteBusyTimeout`.
