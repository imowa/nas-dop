# Reverse proxy (HTTPS) for FileBrowser

If customers open share links over the internet, use **HTTPS** so links are secure. Put FileBrowser behind a reverse proxy with TLS.

## Caddy (minimal example)

Caddy can obtain a certificate automatically (e.g. Let’s Encrypt). Replace `your-domain.com` and `filebrowser-host:10180` with your values.

```caddyfile
your-domain.com {
    reverse_proxy filebrowser-host:10180
}
```

- Share links: `https://your-domain.com/share/<token>` and `https://your-domain.com/dl/<token>` must work **without** auth so public users can access. Caddy above forwards all paths to FileBrowser; no extra auth.
- If you add forward auth (e.g. Authelia) later, allow `/share/*` and `/dl/*` (and `/api/*`, `/static/*` if needed) to bypass auth so share links stay public.

## Nginx (minimal example)

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;
    ssl_certificate     /path/to/fullchain.pem;
    ssl_certificate_key /path/to/privkey.pem;

    location / {
        proxy_pass http://filebrowser-host:10180;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Again, ensure `/share/*` and `/dl/*` are not protected by auth if you add auth elsewhere, so public share links work.
