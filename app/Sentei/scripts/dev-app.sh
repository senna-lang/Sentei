#!/usr/bin/env bash
# swift build → 最小 .app bundle 化 → open で起動。
# `swift run` だと UNUserNotificationCenter が使えないので、通知経路を
# 実地検証する時はこのスクリプトを使う。
#
# Usage:
#   ./scripts/dev-app.sh           # debug build
#   ./scripts/dev-app.sh --release # release build
set -euo pipefail

cd "$(dirname "$0")/.."

CONFIG="debug"
if [[ "${1:-}" == "--release" ]]; then
  CONFIG="release"
fi

echo "==> swift build ($CONFIG)"
swift build -c "$CONFIG"

BIN=".build/$CONFIG/Sentei"
APP=".build/Sentei.app"

if [[ ! -x "$BIN" ]]; then
  echo "ビルド成果物が見つかりません: $BIN" >&2
  exit 1
fi

echo "==> packaging .app bundle at $APP"
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS"
cp "$BIN" "$APP/Contents/MacOS/Sentei"

cat > "$APP/Contents/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleExecutable</key>
  <string>Sentei</string>
  <key>CFBundleIdentifier</key>
  <string>com.sentei.app.dev</string>
  <key>CFBundleName</key>
  <string>Sentei (dev)</string>
  <key>CFBundleDisplayName</key>
  <string>Sentei (dev)</string>
  <key>CFBundleShortVersionString</key>
  <string>0.1.0-dev</string>
  <key>CFBundleVersion</key>
  <string>1</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>LSMinimumSystemVersion</key>
  <string>13.0</string>
  <key>LSUIElement</key>
  <true/>
  <key>NSHumanReadableCopyright</key>
  <string>Sentei dev build</string>
</dict>
</plist>
PLIST

# 既に同名アプリが起動していたら終了させる (通知重複 / PermissionDialog 混乱回避)
pkill -f "$APP/Contents/MacOS/Sentei" 2>/dev/null || true

echo "==> opening $APP"
open "$APP"

echo "done. logs: ~/Library/Logs/Sentei/ (存在すれば), or Console.app で 'Sentei (dev)' を検索"
