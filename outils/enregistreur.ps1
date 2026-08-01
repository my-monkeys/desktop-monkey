# Enregistre le singe sur le vrai bureau Windows, cadre sur lui.
#
# La fenetre du singe (classe "SingeDeBureauClasse") le suit : on lit sa
# position a chaque image et on capture une region centree dessus. Le curseur
# materiel n'apparait pas dans CopyFromScreen : on le redessine a la main.
# Selon le scenario, on pilote aussi la souris pour provoquer le comportement.
#
#   .\enregistreur.ps1 -Scenario suivre -Secondes 6 -Sortie C:\...\frames
param(
    [string]$Scenario = 'suivre',
    [int]$Secondes = 6,
    [int]$Fps = 10,
    [int]$W = 480,
    [int]$H = 340,
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
  [DllImport("user32.dll")] public static extern bool DrawIconEx(IntPtr hdc, int x, int y, IntPtr ic, int w, int h, int step, IntPtr br, int flags);
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

# trouve la fenetre du singe par enumeration (FindWindow s'est montre inefficace
# ici) ; on lit la classe en Unicode et on retient le handle via une portee script
$script:mono = [IntPtr]::Zero
$chercher = [Nat+EnumProc] {
    param($hh, $l)
    $sb = New-Object System.Text.StringBuilder 256
    [void][Nat]::GetClassNameW($hh, $sb, 256)
    if ($sb.ToString() -eq 'SingeDeBureauClasse') { $script:mono = $hh; return $false }
    return $true
}
for ($i = 0; $i -lt 100 -and $script:mono -eq [IntPtr]::Zero; $i++) {
    [void][Nat]::EnumWindows($chercher, [IntPtr]::Zero)
    if ($script:mono -ne [IntPtr]::Zero) { break }
    Start-Sleep -Milliseconds 100
}
$h = $script:mono
"fenetre -> $h" | Add-Content $journal
if ($h -eq [IntPtr]::Zero) { 'fenetre du singe introuvable' | Add-Content $journal; exit 1 }

$scr = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds
$total = $Secondes * $Fps
$dt = [int](1000 / $Fps)
$plein = New-Object System.Drawing.Size($scr.Width, $scr.Height)
$rects = @()  # position du singe par image ; on recadre cote Mac

function RectSinge {
    $r = New-Object Nat+RECT
    [void][Nat]::GetWindowRect($h, [ref]$r)
    return $r
}

for ($f = 0; $f -lt $total; $f++) {
    $t = $f / $Fps

    # --- scenario souris ---
    switch ($Scenario) {
        'suivre' {
            $mx = [int]($scr.Width / 2 + ($scr.Width * 0.32) * [Math]::Sin($t * 1.1))
            $my = [int]($scr.Height * 0.52 + 60 * [Math]::Sin($t * 0.5))
            [void][Nat]::SetCursorPos($mx, $my)
        }
        'degats' {
            # le sprite est en bas de la fenetre (420x300) : viser la, pas le
            # centre, sinon le clic tombe a cote et ne compte pas
            $r = RectSinge
            [void][Nat]::SetCursorPos([int](($r.Left + $r.Right) / 2), $r.Bottom - 34)
            if ($f -eq 12 -or $f -eq 24 -or $f -eq 36) {
                [Nat]::mouse_event([Nat]::LDOWN, 0, 0, 0, [IntPtr]::Zero)
                Start-Sleep -Milliseconds 40
                [Nat]::mouse_event([Nat]::LUP, 0, 0, 0, [IntPtr]::Zero)
            }
        }
        default { } # comportement libre (grimpe, pond), pilote par le config
    }

    # capture plein ecran (methode eprouvee ; le Save d'une sous-region
    # declenchait des erreurs GDI+) + note la position du singe
    $r = RectSinge
    $rects += ('{0},{1},{2},{3},{4}' -f $f, $r.Left, $r.Top, $r.Right, $r.Bottom)
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
    Start-Sleep -Milliseconds $dt
}
$rects | Set-Content (Join-Path $Sortie 'rects.csv')
$n = (Get-ChildItem $Sortie -Filter *.png -EA 0).Count
"fini: $n images" | Add-Content $journal
