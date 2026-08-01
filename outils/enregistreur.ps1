# Enregistre le singe sur le vrai bureau Windows, cadre sur lui.
#
# La fenetre du singe (classe "SingeDeBureauClasse") le suit : on lit sa
# position a chaque image et on capture le plein ecran (le Save d'une
# sous-region declenchait des erreurs GDI+). On note aussi la position du
# curseur, redessine ensuite cote Mac. Selon le scenario, on pilote la souris
# pour provoquer le comportement.
#
#   .\enregistreur.ps1 -Scenario suivre -Secondes 6 -Sortie C:\...\frames
param(
    [string]$Scenario = 'suivre',
    [int]$Secondes = 6,
    [int]$Fps = 10,
    [string]$Sortie = 'C:\Users\glance\SingeTest\frames'
)
Add-Type -AssemblyName System.Windows.Forms, System.Drawing
Add-Type @"
using System;
using System.Text;
using System.Runtime.InteropServices;
public class Nat {
  [DllImport("user32.dll")] public static extern bool EnumWindows(EnumProc f, IntPtr l);
  [DllImport("user32.dll", CharSet=CharSet.Unicode)] public static extern int GetClassNameW(IntPtr h, StringBuilder s, int n);
  [DllImport("user32.dll")] public static extern bool GetWindowRect(IntPtr h, out RECT r);
  [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
  [DllImport("user32.dll")] public static extern void mouse_event(uint f, uint dx, uint dy, uint d, IntPtr i);
  [DllImport("user32.dll")] public static extern bool GetCursorInfo(ref CURSORINFO pci);
  public delegate bool EnumProc(IntPtr h, IntPtr l);
  public struct RECT { public int Left, Top, Right, Bottom; }
  public struct POINT { public int x, y; }
  public struct CURSORINFO { public int cbSize; public int flags; public IntPtr hCursor; public POINT ptScreenPos; }
  public const uint LDOWN = 0x0002, LUP = 0x0004;
}
"@

New-Item -ItemType Directory -Force -Path $Sortie | Out-Null
Get-ChildItem $Sortie -Filter *.png -EA 0 | Remove-Item -Force -EA 0
$journal = Join-Path $Sortie 'record.log'
"debut $(Get-Date -Format HH:mm:ss) scenario=$Scenario" | Set-Content $journal

# trouve une fenetre par sa classe (Unicode) via enumeration
function TrouverClasse($classe) {
    $script:trouve = [IntPtr]::Zero
    $cb = [Nat+EnumProc] {
        param($hh, $l)
        $sb = New-Object System.Text.StringBuilder 256
        [void][Nat]::GetClassNameW($hh, $sb, 256)
        if ($sb.ToString() -eq $classe) { $script:trouve = $hh; return $false }
        return $true
    }
    [void][Nat]::EnumWindows($cb, [IntPtr]::Zero)
    return $script:trouve
}

function Rect($h) {
    $r = New-Object Nat+RECT
    [void][Nat]::GetWindowRect($h, [ref]$r)
    return $r
}

# attend la fenetre du singe
$h = [IntPtr]::Zero
for ($i = 0; $i -lt 100 -and $h -eq [IntPtr]::Zero; $i++) {
    $h = TrouverClasse 'SingeDeBureauClasse'
    if ($h -ne [IntPtr]::Zero) { break }
    Start-Sleep -Milliseconds 100
}
"fenetre -> $h" | Add-Content $journal
if ($h -eq [IntPtr]::Zero) { 'fenetre du singe introuvable' | Add-Content $journal; exit 1 }

$scr = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds
$total = $Secondes * $Fps
$dt = [int](1000 / $Fps)
$plein = New-Object System.Drawing.Size($scr.Width, $scr.Height)
$lignes = @()
# position courante de la souris (pour un deplacement fluide vers les cibles)
$sx = [int]($scr.Width / 2); $sy = [int]($scr.Height / 2)
$explose = $false   # scene pond : une crotte a deja ete crevee
$fin = -1           # images restantes avant l'arret (apres l'explosion)

for ($f = 0; $f -lt $total; $f++) {
    $t = $f / $Fps
    $r = Rect $h

    switch ($Scenario) {
        'suivre' {
            $sx = [int]($scr.Width / 2 + ($scr.Width * 0.30) * [Math]::Sin($t * 1.0))
            $sy = [int]($scr.Height * 0.50 + 60 * [Math]::Sin($t * 0.5))
            [void][Nat]::SetCursorPos($sx, $sy)
        }
        'chasse' {
            $sx = [int]($scr.Width / 2 + ($scr.Width * 0.26) * [Math]::Sin($t * 0.55))
            $sy = [int]($scr.Height * 0.50 + 50 * [Math]::Sin($t * 0.7))
            [void][Nat]::SetCursorPos($sx, $sy)
        }
        'degats' {
            # le sprite est en bas de la fenetre 420x300 : viser la
            $sx = [int](($r.Left + $r.Right) / 2); $sy = $r.Bottom - 34
            [void][Nat]::SetCursorPos($sx, $sy)
            if ($f -eq 12 -or $f -eq 24 -or $f -eq 36) {
                [Nat]::mouse_event([Nat]::LDOWN, 0, 0, 0, [IntPtr]::Zero)
                Start-Sleep -Milliseconds 40
                [Nat]::mouse_event([Nat]::LUP, 0, 0, 0, [IntPtr]::Zero)
            }
        }
        'pond' {
            # on attend la premiere crotte, la souris va la crever, on montre
            # l'explosion, puis on arrete : une seule crotte a l'ecran
            if (-not $explose) {
                $c = TrouverClasse 'CrotteDeBureauClasse'
                if ($c -ne [IntPtr]::Zero) {
                    $cr = Rect $c
                    $tx = [int](($cr.Left + $cr.Right) / 2); $ty = [int](($cr.Top + $cr.Bottom) / 2)
                    $sx = [int]($sx + ($tx - $sx) * 0.45); $sy = [int]($sy + ($ty - $sy) * 0.45)
                    [void][Nat]::SetCursorPos($sx, $sy)
                    if ([Math]::Abs($sx - $tx) -lt 10 -and [Math]::Abs($sy - $ty) -lt 10) {
                        [Nat]::mouse_event([Nat]::LDOWN, 0, 0, 0, [IntPtr]::Zero)
                        Start-Sleep -Milliseconds 40
                        [Nat]::mouse_event([Nat]::LUP, 0, 0, 0, [IntPtr]::Zero)
                        $explose = $true
                        $fin = [int]($Fps * 0.9)  # on filme encore l'explosion, puis stop
                    }
                }
            }
        }
        default { }
    }

    # position reelle du curseur (a redessiner cote Mac)
    $ci = New-Object Nat+CURSORINFO
    $ci.cbSize = [System.Runtime.InteropServices.Marshal]::SizeOf($ci)
    [void][Nat]::GetCursorInfo([ref]$ci)
    $lignes += ('{0},{1},{2},{3},{4},{5},{6}' -f $f, $r.Left, $r.Top, $r.Right, $r.Bottom, $ci.ptScreenPos.x, $ci.ptScreenPos.y)

    try {
        $bmp = New-Object System.Drawing.Bitmap $scr.Width, $scr.Height
        $g = [System.Drawing.Graphics]::FromImage($bmp)
        $g.CopyFromScreen(0, 0, 0, 0, $plein)
        $g.Dispose(); $g = $null
        $bmp.Save((Join-Path $Sortie ('f{0:d3}.png' -f $f)), [System.Drawing.Imaging.ImageFormat]::Png)
    } catch {
        if ($f -lt 3) { "ERREUR image ${f}: $_" | Add-Content $journal }
    } finally {
        if ($g) { $g.Dispose() }
        if ($bmp) { $bmp.Dispose() }
    }
    # scene pond : on coupe peu apres l'explosion, pour ne montrer qu'une crotte
    if ($fin -ge 0) { $fin--; if ($fin -lt 0) { break } }
    Start-Sleep -Milliseconds $dt
}
$lignes | Set-Content (Join-Path $Sortie 'rects.csv')
$n = (Get-ChildItem $Sortie -Filter *.png -EA 0).Count
"fini: $n images" | Add-Content $journal
