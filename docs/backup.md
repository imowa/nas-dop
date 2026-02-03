# Backup

## What to back up

1. **FileBrowser database and config** – Stores users, shares, and settings.
   - **Location (CasaOS)**: `/DATA/AppData/<AppID>/db` (the path mounted as `/db` in the container).
   - Back up the whole `db` folder (e.g. tar or copy to another disk).

2. **Shared files (your photos)** – The path mounted as `/srv` in the container.
   - **Location (CasaOS)**: Usually `/DATA` or your USB path (e.g. `/mnt/usb/share`).
   - Back up this folder regularly (rsync, tar, or your preferred method).

3. **Branding (optional)** – Custom UI files.
   - **Location (CasaOS)**: `/DATA/AppData/<AppID>/branding` (or the path you use in the compose).
   - Back up if you customized `custom.css` or added logo/favicons.

## Recovery

- **Restore db**: Stop the container, replace the `db` folder with your backup, start the container. Shares and user accounts will be restored.
- **Restore files**: Restore the `/srv` source folder from backup. Ensure PUID/PGID ownership is correct.
- **Restore branding**: Copy your backed-up branding folder back to `/DATA/AppData/<AppID>/branding` (or your branding path).
