# Testing Guide for Nas-Dop

## Prerequisites

- Go 1.21 or later installed
- Git (optional, for version control)

## Step 1: Build the Application

```bash
# Navigate to project directory
cd C:\Users\lenov\Nas-Dop

# Download dependencies
go mod tidy

# Build the application
go build -o nas-dop.exe ./cmd/server

# Or run directly without building
go run ./cmd/server
```

## Step 2: Initial Setup

The application will:
1. Create the database at `/data/db/app.sqlite` (or configured DB_PATH)
2. Run migrations to create tables
3. Create a default admin user (username: `admin`, password: `admin`)

**IMPORTANT:** Change the default admin password immediately after first login!

## Step 3: Access the Application

Open your browser and navigate to:
```
http://localhost:8080
```

You should be redirected to the login page.

## Step 4: Test Authentication

1. **Login:**
   - Username: `admin`
   - Password: `admin`
   - Click "Login"
   - Should redirect to `/files`

2. **Logout:**
   - Click "Logout" link
   - Should redirect to `/login`

## Step 5: Test File Operations

### Upload Files
1. Navigate to `/files`
2. Use the upload form to select one or more files
3. Click "Upload"
4. Verify files appear in the file list

### Create Folder
1. Enter a folder name in the "Create Folder" form
2. Click "Create Folder"
3. Verify folder appears in the list with 📁 icon

### Navigate Folders
1. Click on a folder name to navigate into it
2. Verify breadcrumbs show the current path
3. Click breadcrumb links to navigate back

### Download Files
1. Click "Download" link next to a file
2. Verify file downloads correctly

### Delete Files/Folders
1. Click "Delete" button next to a file or folder
2. Confirm the deletion prompt
3. Verify item is removed from the list

### Rename Files/Folders
1. Use the rename functionality (if UI is added)
2. Verify the file/folder is renamed correctly

## Step 6: Test Thumbnails

### Admin Thumbnails
1. Upload image files (JPG, PNG, GIF, or WebP)
2. Navigate to `/files/thumb/{path}` for an image
3. Verify thumbnail is generated and displayed
4. Check that thumbnails are cached in `.thumbcache/` directory
5. Verify subsequent requests are faster (served from cache)

### Thumbnail Cache
1. Upload an image
2. Request its thumbnail
3. Modify the image (re-upload with same name)
4. Request thumbnail again
5. Verify new thumbnail is generated (cache invalidated by mtime)

## Step 7: Test Share Functionality

### Create Share
1. Navigate to a file or folder
2. Click "Share" link
3. Fill in the share form:
   - Name: "Test Share"
   - Password: (optional) "testpass"
   - Expiry: (optional) select a future date
4. Click "Create Share"
5. Copy the share URL displayed

### Access Share (No Password)
1. Open the share URL in an incognito/private window
2. Verify the share page displays with file list
3. Verify share name is displayed
4. Verify files are listed correctly

### Access Share (With Password)
1. Create a password-protected share
2. Open share URL in incognito window
3. Verify password prompt is displayed
4. Enter incorrect password - verify error message
5. Enter correct password - verify access granted
6. Verify password is remembered for the session

## Step 8: Test ZIP Downloads

### Select and Download Files
1. Open a share page with multiple files
2. Check the "Select All" checkbox
3. Verify all file checkboxes are selected
4. Uncheck "Select All" - verify all are deselected
5. Manually select 2-3 files
6. Click "Download Selected as ZIP"
7. Verify ZIP file downloads
8. Extract and verify all selected files are in the ZIP

### ZIP Limits
1. Try to download more than ZipMaxFiles (default 500)
2. Verify appropriate error message
3. Try to download files exceeding ZipMaxBytes (default 2GB)
4. Verify appropriate error message

## Step 9: Test Mobile Responsiveness

1. Open the application on a mobile device or use browser dev tools
2. Verify the UI is responsive and usable on small screens
3. Check that buttons are touch-friendly (min 44px height)
4. Verify tables display correctly on mobile
5. Test all functionality works on mobile

## Step 10: Test Docker Deployment

```bash
# Build and run with Docker Compose
docker-compose up -d

# Check logs
docker-compose logs -f

# Access application
# Navigate to http://localhost:8080

# Stop and remove
docker-compose down
```

## Expected Results

All tests should pass without errors. The application should:
- ✅ Build without compilation errors
- ✅ Start and create database automatically
- ✅ Allow login with default credentials
- ✅ Support all file operations (upload, download, delete, rename, mkdir)
- ✅ Generate thumbnails for images
- ✅ Create and access share links
- ✅ Download multiple files as ZIP
- ✅ Work on mobile devices
- ✅ Run in Docker containers

## Known Limitations

- CSRF protection not yet implemented (Phase 4)
- Rate limiting not yet implemented (Phase 4)
- Rename functionality requires form/UI implementation in templates

