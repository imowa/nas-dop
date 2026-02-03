# Nas-Dop – Studio Photo Sharing on CasaOS

CasaOS-compatible Docker setup for **studio photo sharing**: one folder per customer, share folder link so customers can view and download photos. **Mobile-first UI** – most customers open share links on their phone; the theme is optimized for small screens and touch. GDrive-like, phone-friendly. Targets **Fiberhome HG680FJ** (Armbian + CasaOS from USB).

## Prerequisites

- **Host**: Armbian + CasaOS on HG680FJ (or Raspberry Pi / PC with CasaOS).
- **Docker**: Available from CasaOS.
- **Storage**: USB drive or disk for files (CasaOS often uses `/DATA`).

## Project structure

- **cmd/**, **internal/**, **web/** – From-scratch app (Go backend, SQLite, server-rendered HTML). See **`docs/plan-from-scratch.md`** and **`docs/build-roadmap.md`**.
- **docker/** – Dockerfile and compose (local + CasaOS) for the from-scratch app.
- **config/** – Optional config; `config/example.env` for the app.
- **docs/** – Guides and plans: `docs/storage-setup.md`, `docs/backup.md`, **`docs/plan-from-scratch.md`** (full spec), **`docs/build-roadmap.md`** (phase checklist), **`docs/implementation-status.md`** (what’s done vs TODO).

## Use case: Studio photo sharing

1. Create **one folder per customer** (e.g. `CustomerName`, `CustomerName_OrderID`, or `2025-02-JohnDoe`).
2. Upload that customer’s photos into that folder.
3. Create a **share link** for the folder (admin UI).
4. Send the link to the customer (e.g. by message or email).
5. The customer opens the link (often on their phone) to view and download their photos **without logging in**.

Optional: use a folder naming convention (e.g. customer name + order ID or date) so folders stay organized.

## Install

The app is **built from scratch** (see **Build from scratch** below). When it’s ready:

1. Build and run with Docker: `docker compose -f docker/docker-compose.yml up --build` (local), or use **`docker/docker-compose.casaos.yml`** in CasaOS **Custom App**.
2. **Tutorial:** **[docs/tutorial-install-casaos.md](docs/tutorial-install-casaos.md)** will be adapted for the from-scratch app (CasaOS, port 10180, making share links public).

## Storage

- The app serves files from the volume mapped as **ROOT** (e.g. `/data/files` in CasaOS compose).
- See `config/example.env` and `docker/docker-compose.casaos.yml` for PUID/PGID and paths. Optional: `docs/storage-setup.md` for USB mount.

## First login (when app is ready)

- **URL**: `http://<CasaOS-IP>:10180` (or the port CasaOS assigns).
- **Default credentials**: from env (`DEFAULT_ADMIN_USER` / `DEFAULT_ADMIN_PASSWORD`). **Change after first login.**

## Sharing (public link)

- In the admin UI, pick a **folder** and create a **share**; copy the link. Anyone with the link can **view and download** without logging in.
- URL form: `http://<host>:<port>/share/<token>`. Optional **expiration** or **password** when creating the share.

## Download all (normies-friendly)

Many people don't know how to unzip. Prefer **downloading each photo separately** when you explain to customers:

1. **Download each photo separately** (recommended) – Tap each file to download, or use “Select all” + download if the share UI allows. No unzipping; photos go straight to device. Best for people who don't know how to unzip.
2. **Download as one ZIP file** – On the share page, use “Download folder” (or equivalent). One file; then unzip on computer or phone (see below).

**How to unzip (for customers who chose ZIP):** *Phone:* Tap the downloaded file → Open → Save to Photos or save to Files. *Computer:* Right‑click the .zip file → Extract (or Open with).

Describe both options in simple language so the app is normies-friendly.

## Security

- Change the default admin password after first login.
- For public share links over the internet, put the app behind a **reverse proxy with HTTPS** (e.g. Caddy, Nginx).

## Backup

- Back up the app’s **database** (e.g. `/DATA/AppData/<AppID>/db`) and **file root** (the volume mapped as ROOT). See optional `docs/backup.md`.

## Build from scratch

We can build the same studio photo-sharing app from scratch (Go backend, SQLite, server-rendered HTML, one Docker image) instead of using FileBrowser. The full spec and structure are in **[docs/plan-from-scratch.md](docs/plan-from-scratch.md)**. Use **[docs/build-roadmap.md](docs/build-roadmap.md)** to track phases and checkboxes.

| Phase | Goal |
|-------|------|
| **1** | Backend core: config, SQLite, auth, storage, share store, routes; login, list root, create share, open share link. |
| **2** | Admin file manager: browse any path, upload, mkdir, delete, rename, thumbnails; create share with copy-link. |
| **3** | Share page: checkboxes, “Download selected” (each file) + “Download as ZIP” with “How to unzip?”; thumbnails, password, expiry; mobile-first. |
| **4** | Polish: styling, branding, security (CSRF, rate limit), health endpoint; Docker multi-arch; CasaOS compose and docs. |
| **5** | Optional: JSON API, multiple users, share management. |

The repo is scaffolded with `cmd/`, `internal/`, `web/`, `migrations/`, and `docker/` so implementation can start from Phase 1.

## Note: HG680FJ and 2GB RAM

- The from-scratch app is a single Go binary with low RAM use. Avoid running many heavy containers alongside it.
- If the device struggles, consider enabling swap on Armbian or reducing concurrent usage.
