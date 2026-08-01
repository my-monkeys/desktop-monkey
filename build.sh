#!/usr/bin/env bash
# Compile le singe pour Windows et macOS.
#
#   ./build.sh              les deux plateformes
#   ./build.sh windows      seulement l'executable Windows
#   ./build.sh mac          seulement le binaire macOS
#
# La compilation Windows se fait depuis n'importe quelle machine : pas de cgo,
# donc aucun outil Windows n'est necessaire.
set -euo pipefail
cd "$(dirname "$0")"

LDFLAGS_COMMUNS="-s -w"          # retire la table des symboles : binaire plus leger
mkdir -p dist

construire_windows() {
    echo "→ Windows (amd64)"
    # -H=windowsgui : pas de fenetre de console au lancement
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
        go build -ldflags "$LDFLAGS_COMMUNS -H=windowsgui" -o dist/singe.exe ./cmd/singe
    ls -lh dist/singe.exe | awk '{print "   dist/singe.exe  " $5}'
}

construire_mac() {
    # cgo est obligatoire sur macOS (Cocoa/CoreGraphics), donc compilation sur
    # un Mac uniquement, pour l'architecture courante.
    echo "→ macOS ($(go env GOARCH))"
    CGO_ENABLED=1 go build -ldflags "$LDFLAGS_COMMUNS" -o dist/singe-mac ./cmd/singe
    ls -lh dist/singe-mac | awk '{print "   dist/singe-mac  " $5}'
}

case "${1:-tout}" in
    windows) construire_windows ;;
    mac)     construire_mac ;;
    tout)    construire_windows; construire_mac ;;
    *)       echo "usage: $0 [windows|mac|tout]" >&2; exit 1 ;;
esac
