# Tutorial: Install FileBrowser on CasaOS and Make It Public

This guide walks you through installing the studio photo-sharing app on CasaOS and making FileBrowser public so customers can open share links (and, optionally, so you can reach the admin UI from your network or the internet).

---

## Prerequisites

- **CasaOS** installed on your machine (e.g. Fiberhome HG680FJ with Armbian from USB, or Raspberry Pi / PC).
- **Docker** available (CasaOS uses it by default).
- This repo (Nas-Dop) on your computer or on the CasaOS host, so you can use the compose file and branding folder.

---

## Step 1: Get the app files

You need:

1. The **docker-compose** file: `apps/filebrowser/docker-compose.yml`
2. The **branding** folder (for the mobile-first, GDrive-like theme): `config/filebrowser/branding/` (at least `custom.css`; optional: `img/` for logo and favicons)

If you have the repo on the CasaOS host, you can use it directly. Otherwise, copy the compose file and the whole `config/filebrowser/branding/` folder (including `custom.css` and, if you use it, `img/`) to the host.

---

## Step 2: Install the app in CasaOS

1. Open **CasaOS** in your browser (e.g. `http://<CasaOS-IP>`).
2. Go to **Apps** (or **App Store**).
3. Click **Custom App** (or **Add Custom App** / **Import**).
4. Open `apps/filebrowser/docker-compose.yml` from this repo, select all, and **copy** its contents.
5. **Paste** the contents into the CasaOS Custom App text box.
6. Click **Install** / **Deploy** (or the equivalent button).
7. Wait until CasaOS finishes creating the app. Note the **port** CasaOS shows (often **10180**). If CasaOS asks for an app name, you can use e.g. **FileBrowser** or **Studio Photos**.

After installation, the app appears in your CasaOS app list. You can open it from there (CasaOS will link to `http://<CasaOS-IP>:<port>`).

---

## Step 3: Put the branding folder in place

The custom theme (mobile-first, GDrive-like – optimized for customers on phone) is loaded from a folder that the container sees as `/branding`. CasaOS stores app data under `/DATA/AppData/<AppID>/`. You must put the branding files inside a `branding` folder there.

1. **Find the app data folder**  
   - In CasaOS, open the FileBrowser app’s **settings** or **details** and look for the **data path** or **storage path**, or  
   - On the host (SSH or file manager), list `/DATA/AppData/` and find the folder for this app (often named like `filebrowser` or an ID).

2. **Create the branding folder** (if it doesn’t exist):
   ```bash
   sudo mkdir -p /DATA/AppData/<AppID>/branding
   ```
   Replace `<AppID>` with the actual folder name (e.g. `filebrowser`).

3. **Copy the branding files** from this repo into that folder:
   - Copy **custom.css** from `config/filebrowser/branding/custom.css` to `/DATA/AppData/<AppID>/branding/custom.css`.
   - Optionally copy the **img/** folder (logo, favicons) into `/DATA/AppData/<AppID>/branding/img/`.

   Example (if the app folder is `filebrowser` and you have the repo at `/home/user/Nas-Dop`):
   ```bash
   sudo cp /home/user/Nas-Dop/config/filebrowser/branding/custom.css /DATA/AppData/filebrowser/branding/
   sudo cp -r /home/user/Nas-Dop/config/filebrowser/branding/img /DATA/AppData/filebrowser/branding/  # optional
   ```

4. **Set ownership** so the container can read the files (use the same user/group as CasaOS/Docker, often `1000:1000`):
   ```bash
   sudo chown -R 1000:1000 /DATA/AppData/<AppID>/branding
   ```

5. If the FileBrowser container was already running, **restart the app** from CasaOS so it sees the new files.

---

## Step 4: First login and enable the theme

1. In CasaOS, click the FileBrowser app to open it, or go to **http://&lt;CasaOS-IP&gt;:10180** (use the port CasaOS shows if different).
2. Log in with:
   - **Username:** `admin`  
   - **Password:** `admin`
3. **Change the password** (recommended):
   - Go to **Settings** (gear icon) → **Profile** (or **Account**).
   - Set a new password and save.
4. **Enable the custom theme:**
   - Go to **Settings** → **Global Settings**.
   - Find **Branding directory path** (or **Branding folder**).
   - Set it to: **`/branding`**
   - Save. The page may reload; you should see the mobile-first, GDrive-like theme (optimized for customers on phone).

---

## Step 5: Make FileBrowser “public”

Here “public” means two things: **(A)** share links that anyone with the link can open (no login), and **(B)** (optional) making the admin UI or share links reachable from your network or the internet.

### A. Public share links (customers can open links without login)

This is already how the app is designed:

1. In FileBrowser, create a **folder** (e.g. one per customer: `CustomerName` or `CustomerName_OrderID`).
2. Upload files (e.g. photos) into that folder.
3. **Share the folder:**
   - Right-click the folder (or select it and use the **Share** button / menu).
   - In the share dialog, ensure **Anonymous access** (or “Anyone with the link”) is **enabled** so people can open the link without logging in.
   - Optionally set **expiration** or **password**.
   - Click **Create** / **Generate** and **copy the link**. The link looks like:  
     `http://<CasaOS-IP>:10180/share/<token>`
4. Send that link to your customer. Most will open it on their **phone**; the share page is mobile-first (large tap targets, no horizontal scroll). They can view and download the files **without creating an account or logging in**.

So “making FileBrowser public” for customers = creating shares with anonymous access and sharing those links.

### B. Reachable from your network or the internet

- **From your local network (e.g. home Wi‑Fi)**  
  - Other devices (phones, laptops) can use:
    - Admin UI: `http://<CasaOS-IP>:10180`
    - Share links: `http://<CasaOS-IP>:10180/share/<token>`
  - If it doesn’t work, check that:
    - The device is on the same network as the CasaOS host.
    - No firewall is blocking port **10180** (or the port CasaOS assigned) on the host.

- **From the internet (e.g. for customers outside your home)**  
  - **Option 1 – Port forward (not recommended for admin UI):**  
    Forward port 10180 from your router to the CasaOS host. Then share links would use `http://<your-public-IP>:10180/share/<token>`. Prefer **Option 2** so you can use HTTPS and hide the admin UI.
  - **Option 2 – HTTPS reverse proxy (recommended):**  
    Put FileBrowser behind a reverse proxy with HTTPS (e.g. Caddy or Nginx) and, if you use auth, allow `/share/*` and `/dl/*` without auth so share links stay public. See **docs/reverse-proxy.md** for a minimal Caddy/Nginx example.  
    Then:
    - Admin UI: `https://your-domain.com` (you can protect this with auth).
    - Share links: `https://your-domain.com/share/<token>` (public, no login).

---

## Step 6: Quick test of a public share

1. In FileBrowser, create a folder (e.g. `TestShare`), upload a file, then **Share** the folder with **Anonymous access** on.
2. Copy the share link.
3. Open the link in a **private/incognito** window (or another device). You should see the folder and file **without** being asked to log in.
4. Try the share link on a **phone** first (primary audience): check large tap targets (~44–48px), no horizontal scroll, readable text. Then test on desktop.

---

## Summary

| Goal | What to do |
|------|------------|
| Install on CasaOS | Custom App → paste `apps/filebrowser/docker-compose.yml` → Install. |
| Apply custom theme | Copy `config/filebrowser/branding/` to `/DATA/AppData/<AppID>/branding`, set Branding path to `/branding` in Settings. |
| Public share links | Share a folder → enable Anonymous access → copy link → send to customers. |
| Access from LAN | Use `http://<CasaOS-IP>:10180` and `http://<CasaOS-IP>:10180/share/<token>`; allow port 10180 on firewall if needed. |
| Access from internet with HTTPS | Use a reverse proxy (see docs/reverse-proxy.md); keep `/share/*` and `/dl/*` public. |

For storage (e.g. using a USB drive for files), see **docs/storage-setup.md**. For backups, see **docs/backup.md**.
