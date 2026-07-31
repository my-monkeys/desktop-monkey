#!/usr/bin/env bash
# Compile le singe, l'envoie dans la VM Windows de test, le lance dans la
# session interactive et rapatrie une capture d'ecran.
#
#   ./outils/test-vm.sh              compile, lance, capture apres 8 s
#   ./outils/test-vm.sh 20           idem, capture apres 20 s
#   ./outils/test-vm.sh stop         arrete l'application dans la VM
#
# La VM (dockurr/windows sur cookie-server) est atteinte par l'alias SSH
# "winvm". L'application ne peut pas etre lancee directement depuis SSH : la
# session SSH est la session 0, qui n'a pas de bureau. On passe donc par une
# tache planifiee en ouverture de session interactive.
set -euo pipefail
cd "$(dirname "$0")/.."

HOTE=winvm
DIST=C:/Users/glance/SingeTest
CAPTURES=captures

lancer_dans_session() {
    # $1 = nom de la tache, $2 = executable, $3 = arguments (peut etre vide)
    local arg=""
    [ -n "${3:-}" ] && arg=" -Argument '$3'"
    ssh "$HOTE" "\$a = New-ScheduledTaskAction -Execute '$2'$arg -WorkingDirectory 'C:\\Users\\glance\\SingeTest'
\$p = New-ScheduledTaskPrincipal -UserId \"\$env:COMPUTERNAME\\glance\" -LogonType Interactive -RunLevel Limited
\$s = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit ([TimeSpan]::Zero) -MultipleInstances IgnoreNew
Register-ScheduledTask -TaskName '$1' -Action \$a -Principal \$p -Settings \$s -Force | Out-Null
Start-ScheduledTask -TaskName '$1'" > /dev/null
}

arreter() {
    ssh "$HOTE" "Stop-Process -Name singe -Force -EA 0
Stop-ScheduledTask -TaskName SingeTest -EA 0
Unregister-ScheduledTask -TaskName SingeTest -Confirm:\$false -EA 0
'arrete'" 2>/dev/null | tail -1
}

if [ "${1:-}" = "stop" ]; then arreter; exit 0; fi

ATTENTE="${1:-8}"

echo "→ compilation"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
    go build -ldflags "-s -w -H=windowsgui" -o dist/singe.exe ./cmd/singe

echo "→ envoi vers la VM"
arreter > /dev/null 2>&1 || true
ssh "$HOTE" "New-Item -ItemType Directory -Force -Path $DIST | Out-Null" > /dev/null
scp -q dist/singe.exe "$HOTE:$DIST/singe.exe"
scp -q outils/capture.ps1 "$HOTE:$DIST/capture.ps1"

echo "→ lancement dans la session interactive"
lancer_dans_session SingeTest "$DIST/singe.exe" ""
sleep "$ATTENTE"

echo "→ capture d'ecran"
lancer_dans_session SingeCapture "powershell.exe" "-NoProfile -ExecutionPolicy Bypass -File $DIST/capture.ps1"
sleep 4
ssh "$HOTE" "Unregister-ScheduledTask -TaskName SingeCapture -Confirm:\$false -EA 0" > /dev/null 2>&1 || true

mkdir -p "$CAPTURES"
NOM="$CAPTURES/vm-$(date +%H%M%S).png"
scp -q "$HOTE:$DIST/capture.png" "$NOM"

ssh "$HOTE" "\$p = Get-Process singe -EA 0
if (\$p) { 'application vivante | RAM ' + [math]::Round(\$p.WorkingSet64/1MB,1) + ' Mo' } else { 'APPLICATION ARRETEE' }" 2>/dev/null | tail -1

echo "→ $NOM"
