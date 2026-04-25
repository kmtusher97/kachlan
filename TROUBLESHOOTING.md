# Troubleshooting Guide

## macOS Issues

### ❌ "Apple could not verify 'kachlan' is free of malware"

This is expected! Kachlan is not notarized by Apple because:
- Notarization requires an Apple Developer account ($99/year)
- This is open-source software - you can review the code yourself

**Solution:**
1. Click **"Done"** on the warning
2. Open **System Settings** → **Privacy & Security**
3. Scroll down and click **"Open Anyway"**
4. Click **"Open"** to confirm

After the first launch, macOS will remember your choice.

---

### ❌ "kachlan is damaged and can't be opened"

The app bundle has quarantine attributes.

**Solution:**
```bash
xattr -cr /Applications/kachlan.app
```

Then try opening again using the "Open Anyway" method above.

---

### 🔍 App doesn't appear in Spotlight or Launchpad

macOS needs to rebuild its app database.

**Solution:**
```bash
killall Dock
```

Wait a few seconds for Launchpad to reload.

---

### 🚫 "Operation not permitted" when running commands

Some commands need admin privileges.

**Solution:**
Add `sudo` before the command:
```bash
sudo spctl --add /Applications/kachlan.app
```

Enter your Mac password when prompted.

---

## Windows Issues

### ❌ Windows SmartScreen blocks the app

Windows Defender may show a warning because the app is not signed.

**Solution:**
1. Click **"More info"** on the warning
2. Click **"Run anyway"**

---

## Linux Issues

### ❌ App won't start after installing .deb

The app may not have execute permissions.

**Solution:**
```bash
chmod +x /usr/bin/kachlan
```

---

### 🔍 Can't find app in application menu

The desktop file may need to be refreshed.

**Solution:**
```bash
sudo update-desktop-database
```

Or log out and log back in.

---

## General Issues

### ❌ "ffmpeg not found" error (older versions)

If you're using a version before v0.5.0, you need to install ffmpeg manually.

**Solution - Upgrade:**
Download the latest version from [releases](https://github.com/kmtusher97/kachlan/releases/latest)

**Solution - Manual install:**
- macOS: `brew install ffmpeg`
- Ubuntu: `sudo apt install ffmpeg`
- Windows: `winget install ffmpeg`

---

### 🐛 Found a bug?

Please [open an issue](https://github.com/kmtusher97/kachlan/issues) with:
- Your OS and version
- Steps to reproduce
- Screenshots if applicable

---

### 💬 Need help?

- Check the [README](README.md) for installation instructions
- Search [existing issues](https://github.com/kmtusher97/kachlan/issues)
- Open a [new issue](https://github.com/kmtusher97/kachlan/issues/new)
