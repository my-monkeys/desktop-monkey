#!/usr/bin/env bash
# Fabrique la version macOS distribuable : binaire universel (arm64 + x86_64),
# empaquete dans Desktop Monkey.app, signe avec le certificat Developer ID,
# mis dans un DMG signe. La notarisation Apple s'execute si les identifiants
# App Store Connect sont fournis (ASC_KEY / ASC_KEY_ID / ASC_ISSUER_ID).
#
#   VERSION=2.2.0 ./outils/release-mac.sh
#
# Resultat : dist/DesktopMonkey-<version>.dmg, pret pour la release GitHub et
# le cask Homebrew.
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="${VERSION:-2.2.0}"
ID="${SIGN_ID:-Developer ID Application: Maxim Costa (5C67TFSJ2B)}"
APPNAME="Desktop Monkey"
BUNDLE_ID="fr.mymonkey.desktop-monkey"
EXE="desktop-monkey"

# build avec le Xcode stable : le lipo de Xcode-beta est plus capricieux
if [ -d /Applications/Xcode.app ]; then
  export DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer
fi

echo "== 1. binaire universel =="
CGO_ENABLED=1 GOARCH=arm64 go build -trimpath -o dist/mac-arm64 ./cmd/singe
CGO_ENABLED=1 GOARCH=amd64 go build -trimpath -o dist/mac-amd64 ./cmd/singe
lipo -create dist/mac-arm64 dist/mac-amd64 -output dist/mac-universal
lipo -info dist/mac-universal

echo "== 2. icone =="
rm -rf dist/AppIcon.iconset
python3 outils/icone.py dist/AppIcon.iconset >/dev/null
iconutil -c icns dist/AppIcon.iconset -o dist/AppIcon.icns

echo "== 3. bundle .app =="
APP="dist/$APPNAME.app"
rm -rf "$APP"
mkdir -p "$APP/Contents/MacOS" "$APP/Contents/Resources"
cp dist/mac-universal "$APP/Contents/MacOS/$EXE"
chmod +x "$APP/Contents/MacOS/$EXE"
cp dist/AppIcon.icns "$APP/Contents/Resources/AppIcon.icns"
cat > "$APP/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleName</key><string>$APPNAME</string>
  <key>CFBundleDisplayName</key><string>$APPNAME</string>
  <key>CFBundleIdentifier</key><string>$BUNDLE_ID</string>
  <key>CFBundleExecutable</key><string>$EXE</string>
  <key>CFBundleIconFile</key><string>AppIcon</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>$VERSION</string>
  <key>CFBundleVersion</key><string>$VERSION</string>
  <key>LSMinimumSystemVersion</key><string>11.0</string>
  <key>LSUIElement</key><true/>
  <key>NSHighResolutionCapable</key><true/>
  <key>NSHumanReadableCopyright</key><string>MIT — the monkey belongs to everyone</string>
</dict>
</plist>
PLIST

echo "== 4. signature (hardened runtime + horodatage) =="
codesign --force --options runtime --timestamp --sign "$ID" "$APP/Contents/MacOS/$EXE"
codesign --force --options runtime --timestamp --sign "$ID" "$APP"
codesign --verify --deep --strict --verbose=2 "$APP"

echo "== 5. DMG =="
DMG="dist/DesktopMonkey-$VERSION.dmg"
rm -f "$DMG"
STAGE="$(mktemp -d)"
cp -R "$APP" "$STAGE/$APPNAME.app"
ln -s /Applications "$STAGE/Applications"
hdiutil create -volname "$APPNAME" -srcfolder "$STAGE" -ov -format UDZO "$DMG" >/dev/null
codesign --force --timestamp --sign "$ID" "$DMG"
rm -rf "$STAGE"

echo "== 6. notarisation =="
# Identifiants App Store Connect : variables d'env, sinon valeurs locales par
# defaut (cles dans ~/.appstoreconnect/, hors de tout depot). L'Issuer ID n'est
# jamais dans le repo : il vit dans ~/.appstoreconnect/issuer_id.txt.
DEFKEY="$HOME/.appstoreconnect/private_keys/AuthKey_4SD3G5C575.p8"
[ -z "${ASC_KEY:-}" ] && [ -f "$DEFKEY" ] && ASC_KEY="$DEFKEY"
[ -z "${ASC_KEY_ID:-}" ] && ASC_KEY_ID="4SD3G5C575"
[ -z "${ASC_ISSUER_ID:-}" ] && [ -f "$HOME/.appstoreconnect/issuer_id.txt" ] \
  && ASC_ISSUER_ID="$(cat "$HOME/.appstoreconnect/issuer_id.txt")"
if [ -n "${ASC_KEY:-}" ] && [ -n "${ASC_KEY_ID:-}" ] && [ -n "${ASC_ISSUER_ID:-}" ]; then
  xcrun notarytool submit "$DMG" \
    --key "$ASC_KEY" --key-id "$ASC_KEY_ID" --issuer "$ASC_ISSUER_ID" --wait
  xcrun stapler staple "$DMG"
  spctl -a -vvv -t install "$DMG" || true
else
  echo "  (identifiants ASC absents : notarisation sautee — le DMG est signe"
  echo "   mais pas notarise. Relancer avec ASC_KEY / ASC_KEY_ID / ASC_ISSUER_ID.)"
fi

echo
echo "== fait =="
echo "DMG : $DMG"
shasum -a 256 "$DMG"
