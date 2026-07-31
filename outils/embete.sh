#!/usr/bin/env bash
# Compile la fenetre « embete », l'installe sur un poste Windows et la lance
# dans la session interactive.
#
#   ./outils/embete.sh lancer               compile, envoie et demarre
#   ./outils/embete.sh lancer -max 3        idem, en limitant a 3 fenetres
#   ./outils/embete.sh capture [nom]        rapatrie une capture d'ecran
#   ./outils/embete.sh clic X Y             clique aux coordonnees donnees
#   ./outils/embete.sh bouger X Y           deplace la souris sans cliquer
#   ./outils/embete.sh etat                 dit si l'application tourne
#   ./outils/embete.sh stop                 arrete tout et retire la tache
#
# La machine visee se choisit avec les variables HOTE et UTILISATEUR :
#
#   HOTE=runte-win UTILISATEUR=runte ./outils/embete.sh lancer
#
# L'application ne peut pas etre lancee directement depuis SSH : la session SSH
# est la session 0, qui n'a pas de bureau. Tout ce qui doit s'afficher passe
# donc par une tache planifiee en session interactive.
set -euo pipefail
cd "$(dirname "$0")/.."

HOTE="${HOTE:-winvm}"
UTILISATEUR="${UTILISATEUR:-glance}"
DIST="C:/Users/$UTILISATEUR/Embete"
DOSSIER="C:\\Users\\$UTILISATEUR\\Embete"
CAPTURES=captures

# Envoie un script PowerShell par l'entree standard. Le shell par defaut
# d'OpenSSH est cmd.exe sur certains postes et powershell.exe sur d'autres :
# passer le script en argument ne marcherait que sur les seconds.
ps_distant() {
    ssh "$HOTE" "powershell -NoProfile -ExecutionPolicy Bypass -Command -"
}

# Lance un programme dans la session de l'utilisateur.
#   $1 tache, $2 executable, $3 arguments (facultatif)
en_session() {
    local arg=""
    [ -n "${3:-}" ] && arg=" -Argument '$3'"
    ps_distant > /dev/null 2>&1 <<EOF
\$a = New-ScheduledTaskAction -Execute '$2'$arg -WorkingDirectory '$DOSSIER'
\$p = New-ScheduledTaskPrincipal -UserId "\$env:COMPUTERNAME\\$UTILISATEUR" -LogonType Interactive -RunLevel Limited
\$s = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit ([TimeSpan]::Zero) -MultipleInstances IgnoreNew
Register-ScheduledTask -TaskName '$1' -Action \$a -Principal \$p -Settings \$s -Force | Out-Null
Start-ScheduledTask -TaskName '$1'
EOF
}

retirer_tache() {
    ps_distant > /dev/null 2>&1 <<EOF || true
Unregister-ScheduledTask -TaskName '$1' -Confirm:\$false -EA 0
EOF
}

case "${1:-}" in
lancer)
    shift
    echo "→ compilation"
    GOOS=windows GOARCH=amd64 CGO_ENABLED=0 \
        go build -ldflags "-s -w -H=windowsgui" -o dist/embete.exe ./cmd/embete

    echo "→ envoi vers $HOTE"
    ps_distant > /dev/null <<EOF
New-Item -ItemType Directory -Force -Path '$DOSSIER' | Out-Null
EOF
    scp -q dist/embete.exe "$HOTE:$DIST/embete.exe"
    scp -q outils/capture.ps1 "$HOTE:$DIST/capture.ps1"
    scp -q outils/souris.ps1 "$HOTE:$DIST/souris.ps1"

    echo "→ lancement dans la session interactive"
    en_session Embete "$DIST/embete.exe" "$*"
    ;;

capture)
    en_session EmbeteCapture "powershell.exe" \
        "-NoProfile -ExecutionPolicy Bypass -File $DIST/capture.ps1"
    sleep 4
    mkdir -p "$CAPTURES"
    NOM="$CAPTURES/${2:-embete-$(date +%H%M%S)}.png"
    scp -q "$HOTE:$DIST/capture.png" "$NOM"
    retirer_tache EmbeteCapture
    echo "→ $NOM"
    ;;

clic | bouger)
    en_session EmbeteSouris "powershell.exe" \
        "-NoProfile -ExecutionPolicy Bypass -File $DIST/souris.ps1 -Action $1 -X $2 -Y $3"
    sleep 2
    retirer_tache EmbeteSouris
    ;;

etat)
    ps_distant 2>/dev/null <<'EOF' | tail -1
$p = Get-Process embete -EA 0
if ($p) { 'en marche | ' + $p.MainWindowTitle } else { 'arretee' }
EOF
    ;;

stop)
    ps_distant 2>/dev/null <<'EOF' | tail -1
Stop-Process -Name embete -Force -EA 0
foreach ($t in 'Embete', 'EmbeteCapture', 'EmbeteSouris') {
    Unregister-ScheduledTask -TaskName $t -Confirm:$false -EA 0
}
'arretee'
EOF
    ;;

*)
    echo "usage: $0 <lancer|capture|clic X Y|bouger X Y|etat|stop>" >&2
    exit 1
    ;;
esac
