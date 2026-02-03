# Storage setup (HG680FJ / CasaOS)

## CasaOS and /DATA

- CasaOS typically uses **`/DATA`** for app data. It may be on the USB drive if you configured it that way during CasaOS setup.
- FileBrowser’s compose maps:
  - `/DATA/AppData/$AppID/db` → container `/db` (database)
  - `/DATA` → container `/srv` (files you share)
  - `/DATA/AppData/$AppID/branding` → container `/branding` (custom UI)

## Using a dedicated folder on USB for shared files

If you want shared files to live on a specific USB partition (e.g. a large drive):

1. **Mount the USB partition** (e.g. `/dev/sda1`) at a path like `/mnt/usb`:
   ```bash
   sudo mkdir -p /mnt/usb
   sudo mount /dev/sda1 /mnt/usb
   ```
   To mount automatically on boot, add an entry to `/etc/fstab`.

2. **Create a folder** for shared files, e.g. `/mnt/usb/share`.

3. **Set ownership** so FileBrowser (running as PUID/PGID) can read/write:
   ```bash
   sudo chown -R 1000:1000 /mnt/usb/share
   ```
   Use your actual PUID/PGID (run `id` to get them).

4. **In the compose**, change the second volume from:
   ```yaml
   - type: bind
     source: /DATA
     target: /srv
   ```
   to:
   ```yaml
   - type: bind
     source: /mnt/usb/share
     target: /srv
   ```

## Branding folder when CasaOS runs the stack

CasaOS may run the compose from a different working directory, so a relative path like `./branding` in the compose can fail. Use one of:

- **Absolute path**: If the host has this repo, set the branding volume source to the absolute path of `config/filebrowser/branding` (e.g. `/home/user/Nas-Dop/config/filebrowser/branding`).
- **CasaOS AppData**: Copy the contents of `config/filebrowser/branding/` from this repo into `/DATA/AppData/<AppID>/branding` on the host. The compose already mounts `/DATA/AppData/$AppID/branding` as `/branding`.

After copying, ensure `custom.css` (and optional `img/`) are inside that folder so FileBrowser can load the theme when you set **Settings → Global Settings → Branding directory path** to `/branding`.
