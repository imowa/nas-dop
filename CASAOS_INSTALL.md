# Installing Nas-Dop on CasaOS - Beginner's Guide

This guide will help you install Nas-Dop (a simple file sharing app) on your CasaOS home server. No technical experience required!

## What You'll Need

- A CasaOS server (already set up and running)
- Access to your CasaOS web interface
- About 10-15 minutes

## What is Nas-Dop?

Nas-Dop is a simple file sharing application that lets you:
- Upload and manage files through a web browser
- Create share links to share files with others
- Download multiple files as a ZIP
- View image thumbnails
- Access everything from your phone or computer

## Installation Methods

There are two ways to install Nas-Dop on CasaOS:

1. **Method A: Using the App Store** (Easier - if available)
2. **Method B: Manual Installation** (Works for everyone)

We'll cover both methods below.

---

## Method A: App Store Installation (If Available)

**Note:** This method only works if Nas-Dop has been added to the CasaOS App Store. If you don't see it in the store, use Method B instead.

### Steps:

1. **Open CasaOS**
   - Open your web browser
   - Go to your CasaOS address (usually `http://casaos.local` or your server's IP address)
   - Log in with your username and password

2. **Open the App Store**
   - Click on the "App Store" icon in the CasaOS interface
   - It looks like a shopping bag or grid of apps

3. **Search for Nas-Dop**
   - Use the search bar at the top
   - Type "Nas-Dop" or "file sharing"

4. **Install the App**
   - Click on the Nas-Dop app card
   - Click the "Install" button
   - Wait for the installation to complete (usually 1-2 minutes)

5. **Open the App**
   - Once installed, click "Open" or find the app in your CasaOS dashboard
   - The app will open in a new tab

6. **Skip to "First-Time Setup" section below**

---

## Method B: Manual Installation (Recommended)

This method works for everyone and gives you full control over the installation.

### Step 1: Get the Files

You need to get the Nas-Dop files onto your CasaOS server. There are two ways to do this:

#### Option 1: Using SSH (If you're comfortable with it)

1. **Connect to your CasaOS server via SSH**
   - Use an SSH client like PuTTY (Windows) or Terminal (Mac/Linux)
   - Connect to your server's IP address
   - Log in with your username and password

2. **Download the files**
   ```bash
   cd /DATA/AppData
   git clone https://github.com/yourusername/nas-dop.git
   cd nas-dop
   ```

#### Option 2: Using CasaOS File Manager (Easier)

1. **Download the project files to your computer**
   - Go to the Nas-Dop GitHub page or download location
   - Click "Code" → "Download ZIP"
   - Extract the ZIP file on your computer

2. **Upload to CasaOS**
   - Open CasaOS in your browser
   - Click on "Files" app
   - Navigate to `/DATA/AppData/`
   - Create a new folder called `nas-dop`
   - Upload all the extracted files into this folder

### Step 2: Build the Docker Image

Now we need to build the Docker image. This is like creating a package that CasaOS can run.

1. **Open CasaOS Terminal**
   - In CasaOS, look for a "Terminal" or "Console" app
   - If you don't have one, you can use SSH instead

2. **Navigate to the project folder**
   ```bash
   cd /DATA/AppData/nas-dop
   ```

3. **Build the Docker image**
   ```bash
   docker build -t nas-dop:latest .
   ```

   This will take 2-5 minutes. You'll see lots of text scrolling by - that's normal!

   Wait until you see "Successfully built" and "Successfully tagged nas-dop:latest"

### Step 3: Install in CasaOS

Now we'll add the app to your CasaOS dashboard.

1. **Open CasaOS App Store**
   - Go back to your CasaOS web interface
   - Click on "App Store"

2. **Import Custom App**
   - Look for a button that says "Install a customized app" or "+" button
   - Click on it

3. **Use the Compose File**
   - Select "Import from docker-compose"
   - Click "Browse" or "Upload"
   - Navigate to `/DATA/AppData/nas-dop/docker-compose.casaos.yml`
   - Select the file and click "Open"

4. **Review Settings** (Optional)
   - You can change the port if 8080 is already in use
   - You can change the data folder location
   - For most users, the defaults are fine

5. **Install**
   - Click "Install" or "Submit"
   - Wait for the installation to complete (30 seconds to 1 minute)

---

## First-Time Setup

After installation, you need to set up your account.

### Step 1: Access the App

1. **Find the app in your CasaOS dashboard**
   - Look for "Nas-Dop" or "Studio Photos" icon
   - Click on it to open

2. **Or access directly**
   - Open your browser
   - Go to: `http://your-casaos-ip:8080`
   - Replace `your-casaos-ip` with your server's IP address
   - Example: `http://192.168.1.100:8080`

### Step 2: Login

1. **Use the default credentials**
   - Username: `admin`
   - Password: `admin`

2. **IMPORTANT: Change your password!**
   - After logging in, you should change the default password
   - (Note: Password change feature may need to be added in future updates)

---

## How to Use Nas-Dop

### Uploading Files

1. **Click the file input** under "Upload Files"
2. **Select one or more files** from your computer
3. **Click "Upload"**
4. Your files will appear in the list

### Creating Folders

1. **Type a folder name** in the "Create Folder" box
2. **Click "Create Folder"**
3. Click on the folder name to open it

### Sharing Files

1. **Click "Share"** next to any file or folder
2. **Fill in the form:**
   - Name: Give your share a friendly name
   - Password: (Optional) Add a password for security
   - Expiry: (Optional) Set when the share expires
3. **Click "Create Share"**
4. **Copy the share link** and send it to anyone
5. They can access the files without logging in!

### Downloading Multiple Files

1. **Open a share link** (or your admin view)
2. **Check the boxes** next to files you want
3. **Click "Download Selected as ZIP"**
4. All selected files will download as one ZIP file

---

## Troubleshooting

### Problem: Can't access the app

**Solution:**
- Check if the container is running in CasaOS dashboard
- Try accessing with your server's IP: `http://192.168.1.X:8080`
- Make sure port 8080 isn't blocked by your firewall
- Try restarting the container

### Problem: Build failed

**Solution:**
- Make sure all files were uploaded correctly
- Check that you're in the right directory: `/DATA/AppData/nas-dop`
- Try running the build command again
- Check CasaOS has enough disk space

### Problem: Files not uploading

**Solution:**
- Check file size limits (default is 100MB per file)
- Make sure you have enough disk space
- Try uploading smaller files first to test
- Check browser console for errors (F12 key)

---

## Tips and Recommendations

### Security Tips

1. **Change the default password** as soon as possible
2. **Use passwords on shares** when sharing sensitive files
3. **Set expiry dates** on shares you don't need long-term
4. **Don't share the admin login** with others

### Performance Tips

1. **Image thumbnails** are automatically generated and cached
2. **ZIP downloads** work best with under 100 files at a time
3. **Large files** may take time to upload depending on your network

### Accessing from Outside Your Home

To access Nas-Dop from outside your home network, you'll need to:
- Set up port forwarding on your router (port 8080)
- Or use CasaOS's built-in remote access features
- Or use a VPN to connect to your home network

---

## Summary

Congratulations! You've successfully installed Nas-Dop on CasaOS. You can now:

✅ Upload and manage files through your browser
✅ Create folders to organize your files
✅ Share files with others using simple links
✅ Download multiple files as ZIP archives
✅ Access everything from any device

**Need Help?**
- Check the TESTING.md file for detailed testing instructions
- Report issues on the GitHub repository
- Check CasaOS community forums for support

Enjoy your new file sharing system! 🎉
