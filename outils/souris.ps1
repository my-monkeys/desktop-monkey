# Pilote la souris de la session interactive, pour eprouver les interactions
# (clic sur le singe, deplacement du cadavre) sans intervention humaine.
#
#   .\souris.ps1 -Action clic     -X 640 -Y 360
#   .\souris.ps1 -Action glisser  -X 640 -Y 360 -VersX 300 -VersY 200
param(
    [ValidateSet('clic', 'glisser', 'bouger')] [string]$Action = 'clic',
    [int]$X, [int]$Y, [int]$VersX, [int]$VersY
)

Add-Type @"
using System;
using System.Runtime.InteropServices;
public class Souris {
    [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
    [DllImport("user32.dll")] public static extern void mouse_event(uint f, uint dx, uint dy, uint d, IntPtr i);
    public const uint GAUCHE_BAS = 0x0002;
    public const uint GAUCHE_HAUT = 0x0004;
}
"@

function Bouger($x, $y) { [Souris]::SetCursorPos($x, $y) | Out-Null; Start-Sleep -Milliseconds 40 }

switch ($Action) {
    'bouger' { Bouger $X $Y }

    'clic' {
        Bouger $X $Y
        [Souris]::mouse_event([Souris]::GAUCHE_BAS, 0, 0, 0, [IntPtr]::Zero)
        Start-Sleep -Milliseconds 60
        [Souris]::mouse_event([Souris]::GAUCHE_HAUT, 0, 0, 0, [IntPtr]::Zero)
    }

    'glisser' {
        Bouger $X $Y
        [Souris]::mouse_event([Souris]::GAUCHE_BAS, 0, 0, 0, [IntPtr]::Zero)
        # deplacement progressif : le programme suit le curseur image par image
        for ($i = 1; $i -le 20; $i++) {
            Bouger ([int]($X + ($VersX - $X) * $i / 20)) ([int]($Y + ($VersY - $Y) * $i / 20))
        }
        Start-Sleep -Milliseconds 100
        [Souris]::mouse_event([Souris]::GAUCHE_HAUT, 0, 0, 0, [IntPtr]::Zero)
    }
}

"$Action fait en ($X,$Y)" | Out-File (Join-Path $PSScriptRoot 'souris.log') -Encoding UTF8 -Append
