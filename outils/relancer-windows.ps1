# Relance le singe dans la session bureau d'une machine Windows distante.
#
#   scp outils/relancer-windows.ps1 <hote>:"C:/Users/<user>/relancer.ps1"
#   ssh <hote> "powershell -NoProfile -ExecutionPolicy Bypass -File C:\Users\<user>\relancer.ps1 -Exe C:\chemin\singe.exe"
#
# Une session SSH vit en session 0 : un programme qu'elle lance n'apparait sur
# aucun ecran. Pour le faire naitre dans la session ouverte de l'utilisateur, il
# faut passer par le planificateur de taches avec un jeton interactif.
#
# Le piege qui coute des heures : sur un portable **sur batterie**, le
# planificateur refuse de demarrer la tache — et n'en dit rien. schtasks /run
# repond « operation reussie », la tache s'acheve avec le code 0, et aucun
# processus n'apparait. Il faut lui donner AllowStartIfOnBatteries.
param(
    [Parameter(Mandatory = $true)][string]$Exe,
    [string]$Tache = "SingeRelance"
)
$ErrorActionPreference = "Stop"

$sid = ([System.Security.Principal.WindowsIdentity]::GetCurrent()).User.Value

$action    = New-ScheduledTaskAction -Execute $Exe
$principal = New-ScheduledTaskPrincipal -UserId $sid -LogonType Interactive
$settings  = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries `
                                          -DontStopIfGoingOnBatteries `
                                          -ExecutionTimeLimit ([TimeSpan]::Zero) `
                                          -MultipleInstances IgnoreNew

Register-ScheduledTask -TaskName $Tache -Action $action -Principal $principal `
                       -Settings $settings -Force | Out-Null
Start-ScheduledTask -TaskName $Tache
Start-Sleep -Seconds 5

# la tache n'a servi qu'a le mettre au monde : le processus lui survit
Unregister-ScheduledTask -TaskName $Tache -Confirm:$false

$nom = [System.IO.Path]::GetFileNameWithoutExtension($Exe)
$p = Get-Process $nom -ErrorAction SilentlyContinue
if ($p) { "lance : pid $($p.Id), session $($p.SessionId)" } else { throw "rien n'a demarre" }
