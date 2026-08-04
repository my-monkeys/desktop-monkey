#!/usr/bin/env bash
# Fabrique l'executable Windows distribuable.
#
#   VERSION=2.6.0 ./outils/release-windows.sh
#
# Resultat : dist/monkey.exe, l'asset attendu par l'auto-mise a jour
# (internal/maj le cherche sous ce nom exact dans la derniere release).
#
# Les trois options du lieur comptent, et l'oubli de la premiere ne se voit
# qu'a l'usage :
#
#   -H windowsgui  sans elle, l'application est un programme de console :
#                  une fenetre noire s'ouvre a cote du singe, et la fermer
#                  le tue. C'est arrive en 2.6.0.
#   -s -w          retire les tables de symboles : l'executable passe de 17 a
#                  13 Mo, autant de moins a telecharger a chaque mise a jour.
#   -X main.version   sans elle la version reste "dev", que l'auto-mise a jour
#                  laisse tranquille — le singe ne se mettrait plus a jour.
#
# L'icone de l'executable, elle, vient des .syso commites dans cmd/singe
# (voir outils/icone.py) : le lieur les ramasse tout seul.
set -euo pipefail
cd "$(dirname "$0")/.."

VERSION="${VERSION:?VERSION=x.y.z requis}"
mkdir -p dist

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
  -ldflags "-H windowsgui -s -w -X main.version=$VERSION" \
  -o dist/monkey.exe ./cmd/singe

ls -l dist/monkey.exe
echo "fait : dist/monkey.exe (version $VERSION)"
