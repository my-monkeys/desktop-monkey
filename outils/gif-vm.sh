#!/usr/bin/env bash
# Enregistre une scene du singe sur la VM Windows, cadree sur lui, et rapatrie
# les images (a assembler en GIF cote Mac).
#
#   ./outils/gif-vm.sh suivre 6      scene "suivre", 6 secondes
#   ./outils/gif-vm.sh degats 7
#   ./outils/gif-vm.sh grimpe 9
#   ./outils/gif-vm.sh pond 9
#   ./outils/gif-vm.sh stop          arrete tout
#
# Les images arrivent dans captures/<scene>/f###.png.
set -euo pipefail
cd "$(dirname "$0")/.."

HOTE=winvm
DIST=C:/Users/glance/SingeTest
FRAMES="$DIST/frames"
CFG='C:/Users/glance/AppData/Roaming/SingeDeBureau/config.json'

# PowerShell sur stdin : forme portable (voir memoire machines-windows)
psh() { ssh "$HOTE" "powershell -NoProfile -Command -"; }

taches_off() {
    psh <<'EOF' >/dev/null 2>&1 || true
Stop-Process -Name singe -Force -EA 0
foreach ($t in 'SingeTest','SingeRecord') {
  Stop-ScheduledTask -TaskName $t -EA 0
  Unregister-ScheduledTask -TaskName $t -Confirm:$false -EA 0
}
'ok'
EOF
}

lancer() { # $1 nom tache, $2 exe, $3 arguments
    local arg=""
    [ -n "${3:-}" ] && arg=" -Argument '$3'"
    psh <<EOF >/dev/null
\$a = New-ScheduledTaskAction -Execute '$2'$arg -WorkingDirectory 'C:\\Users\\glance\\SingeTest'
\$p = New-ScheduledTaskPrincipal -UserId "\$env:COMPUTERNAME\\glance" -LogonType Interactive -RunLevel Limited
\$s = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit ([TimeSpan]::Zero) -MultipleInstances IgnoreNew
Register-ScheduledTask -TaskName '$1' -Action \$a -Principal \$p -Settings \$s -Force | Out-Null
Start-ScheduledTask -TaskName '$1'
EOF
}

if [ "${1:-}" = "stop" ]; then taches_off; echo "arrete"; exit 0; fi

SCENE="${1:-suivre}"
SECS="${2:-6}"

# config qui force le comportement voulu, sans bruit (pas de bulles)
case "$SCENE" in
  suivre) KV='"chance_ami":1,"chance_chasse":0' ;;
  chasse) KV='"chance_ami":1,"chance_chasse":1,"chance_vol":0' ;;
  degats) KV='"chance_ami":0,"coeurs":3' ;;
  grimpe) KV='"chance_ami":0,"chance_grimpe":1,"secondes_avant_vie_seule":1' ;;
  pond)   KV='"chance_ami":0,"chance_crotte":0.75,"secondes_avant_vie_seule":1' ;;
  *) echo "scene inconnue: $SCENE"; exit 1 ;;
esac
CONFIG="{\"planche\":\"assets/singe2.json\",\"chance_repas\":0,\"chance_jeu\":0,\"chance_grimpe\":0,\"chance_crotte\":0,\"parle\":false,$KV}"

echo "→ compilation"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags "-s -w -H=windowsgui" -o dist/singe.exe ./cmd/singe

echo "→ envoi (exe, enregistreur, config)"
taches_off
psh <<EOF >/dev/null
New-Item -ItemType Directory -Force -Path '$DIST' | Out-Null
New-Item -ItemType Directory -Force -Path (Split-Path '$CFG') | Out-Null
Set-Content -Path '$CFG' -Value '$CONFIG' -Encoding utf8
EOF
scp -q dist/singe.exe "$HOTE:$DIST/singe.exe"
scp -q outils/enregistreur.ps1 "$HOTE:$DIST/enregistreur.ps1"

echo "→ lancement du singe"
lancer SingeTest "$DIST/singe.exe" ""
# attend que le processus soit bien vivant (jusqu'a ~12 s)
vivant=0
for i in $(seq 1 12); do
    n=$(ssh "$HOTE" "(Get-Process singe -EA 0 | Measure-Object).Count" 2>/dev/null | tr -dc '0-9')
    if [ "${n:-0}" != "0" ]; then vivant=1; echo "  singe vivant apres ${i}s"; break; fi
    sleep 1
done
[ "$vivant" = "0" ] && echo "  ATTENTION: le singe ne semble pas lance"
sleep 2

echo "→ enregistrement ($SCENE, ${SECS}s)"
lancer SingeRecord "powershell.exe" "-NoProfile -WindowStyle Hidden -ExecutionPolicy Bypass -File $DIST/enregistreur.ps1 -Scenario $SCENE -Secondes $SECS -Sortie $FRAMES"
sleep $((SECS + 8))

echo "→ rapatriement des images"
LOCAL="captures/$SCENE"
rm -rf "$LOCAL"; mkdir -p "$LOCAL"
scp -q "$HOTE:$FRAMES/*.png" "$LOCAL/" 2>/dev/null || echo "  (aucune image — voir plus haut)"
scp -q "$HOTE:$FRAMES/rects.csv" "$LOCAL/" 2>/dev/null || true
N=$(ls "$LOCAL"/f*.png 2>/dev/null | wc -l | tr -d ' ')
echo "→ $N images dans $LOCAL"

taches_off >/dev/null
