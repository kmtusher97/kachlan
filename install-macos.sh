#!/bin/bash
# Kachlan macOS Installation Helper
# This script removes quarantine attributes and installs the app

set -e

echo "🍋 Kachlan macOS Installation Helper"
echo "===================================="
echo ""

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "❌ This script is for macOS only"
    exit 1
fi

# Find the DMG file
DMG_FILE=""
if [ -f "$HOME/Downloads/kachlan-gui_darwin_universal.dmg" ]; then
    DMG_FILE="$HOME/Downloads/kachlan-gui_darwin_universal.dmg"
elif [ $# -eq 1 ] && [ -f "$1" ]; then
    DMG_FILE="$1"
else
    echo "❌ Could not find kachlan DMG file in ~/Downloads/"
    echo ""
    echo "Usage: $0 [path-to-dmg]"
    echo "Example: $0 ~/Downloads/kachlan-gui_darwin_universal.dmg"
    exit 1
fi

echo "📦 Found DMG: $DMG_FILE"
echo ""

# Remove quarantine from DMG
echo "🔓 Removing quarantine attribute from DMG..."
xattr -d com.apple.quarantine "$DMG_FILE" 2>/dev/null || true
echo "✅ DMG quarantine removed"
echo ""

# Mount the DMG
echo "💿 Mounting DMG..."
MOUNT_POINT=$(hdiutil attach "$DMG_FILE" | grep "/Volumes" | awk '{print $3}')

if [ -z "$MOUNT_POINT" ]; then
    echo "❌ Failed to mount DMG"
    exit 1
fi

echo "✅ Mounted at: $MOUNT_POINT"
echo ""

# Copy app to Applications
echo "📂 Installing to /Applications..."
rm -rf /Applications/kachlan.app 2>/dev/null || true
cp -R "$MOUNT_POINT/kachlan.app" /Applications/

# Unmount DMG
echo "💿 Unmounting DMG..."
hdiutil detach "$MOUNT_POINT" -quiet

# Remove quarantine from installed app
echo "🔓 Removing quarantine from installed app..."
xattr -cr /Applications/kachlan.app
echo "✅ App quarantine removed"
echo ""

# Refresh Launchpad
echo "🔄 Refreshing Launchpad..."
killall Dock
echo "✅ Launchpad refreshed"
echo ""

echo "✨ Installation complete!"
echo ""
echo "⚠️  IMPORTANT - First Launch Instructions:"
echo ""
echo "macOS will block the app because it's not notarized by Apple."
echo "This is normal for open-source apps."
echo ""
echo "To bypass Gatekeeper:"
echo ""
echo "Method 1 (Easiest):"
echo "  1. Try to open kachlan from Launchpad or Applications"
echo "  2. Click 'Done' on the warning"
echo "  3. Open System Settings → Privacy & Security"
echo "  4. Click 'Open Anyway' button"
echo "  5. Click 'Open' in the confirmation"
echo ""
echo "Method 2 (Quick):"
echo "  Right-click kachlan in Applications → 'Open' → click 'Open'"
echo ""
echo "After the first launch, the app will open normally!"
echo ""
