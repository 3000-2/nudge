#!/bin/sh
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OUT_DIR="${1:-$PROJECT_ROOT/build}"
ARCH="${2:-$(uname -m)}"

# Map arch names
case "$ARCH" in
  arm64)  TARGET="arm64-apple-macosx12.0" ;;
  amd64)  TARGET="x86_64-apple-macosx12.0" ;;
  x86_64) TARGET="x86_64-apple-macosx12.0" ;;
  universal) ;;
  *)      echo "unsupported arch: $ARCH" >&2; exit 1 ;;
esac

APP_DIR="$OUT_DIR/Nudge.app"
CONTENTS="$APP_DIR/Contents"
MACOS="$CONTENTS/MacOS"
RESOURCES="$CONTENTS/Resources"

echo "▸ Building Nudge.app ($ARCH)..."

# Clean
rm -rf "$APP_DIR"
mkdir -p "$MACOS" "$RESOURCES"

# Compile
SDK=$(xcrun --show-sdk-path)

if [ "$ARCH" = "universal" ]; then
  TMPDIR_BUILD=$(mktemp -d)
  trap 'rm -rf "$TMPDIR_BUILD"' EXIT

  xcrun swiftc -O -target arm64-apple-macosx12.0 -sdk "$SDK" \
    -framework UserNotifications -framework Foundation \
    "$SCRIPT_DIR/main.swift" -o "$TMPDIR_BUILD/Nudge_arm64"

  xcrun swiftc -O -target x86_64-apple-macosx12.0 -sdk "$SDK" \
    -framework UserNotifications -framework Foundation \
    "$SCRIPT_DIR/main.swift" -o "$TMPDIR_BUILD/Nudge_x86_64"

  lipo -create \
    "$TMPDIR_BUILD/Nudge_arm64" \
    "$TMPDIR_BUILD/Nudge_x86_64" \
    -output "$MACOS/Nudge"
else
  xcrun swiftc -O -target "$TARGET" -sdk "$SDK" \
    -framework UserNotifications -framework Foundation \
    "$SCRIPT_DIR/main.swift" -o "$MACOS/Nudge"
fi

chmod +x "$MACOS/Nudge"

# Assemble bundle
cp "$SCRIPT_DIR/Info.plist" "$CONTENTS/Info.plist"
cp "$SCRIPT_DIR/Assets/nudge.icns" "$RESOURCES/nudge.icns"

# Ad-hoc sign (required for UNUserNotificationCenter)
codesign --force --sign - "$APP_DIR"

echo "✓ Nudge.app built at $APP_DIR"
