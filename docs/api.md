# API outline

For future SPA or integrations. Full route table and behavior are in **[plan-from-scratch.md](plan-from-scratch.md)** §6.

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/` | No | Redirect to login or /files |
| GET/POST | `/login` | No | Login form / authenticate |
| POST | `/logout` | Yes | Clear session |
| GET | `/files`, `/files/*path` | Yes | List or download (?dl=1) |
| POST/DELETE/PATCH | `/files/*path` | Yes | Upload, mkdir, delete, rename |
| GET/POST | `/share/new` | Yes | Create share form / submit |
| GET/POST | `/share/<token>` | No | Public share page / password |
| GET | `/share/<token>/dl/*path` | No | Download single file |
| GET | `/share/<token>/zip?paths=...` | No | Stream ZIP of selected files |
| GET | `/files/thumb/*path` | Yes | Thumbnail (admin) |
| GET | `/share/<token>/thumb/*path` | No | Thumbnail (share) |
| GET | `/health` | No | Health check (200 OK) |

Optional later: JSON API under `/api/` (same logic, JSON responses).
