# Etape 2 : deplie le debordement, clic droit sur l'icone du singe, capture le
# menu ouvert (verification des jauges d'humeur).
Add-Type -AssemblyName System.Windows.Forms, System.Drawing
Add-Type @"
using System;
using System.Runtime.InteropServices;
public class SourisD {
  [DllImport("user32.dll")] public static extern bool SetCursorPos(int x, int y);
  [DllImport("user32.dll")] public static extern void mouse_event(uint f, uint dx, uint dy, uint d, IntPtr i);
  public const uint LDOWN = 0x0002, LUP = 0x0004, RDOWN = 0x0008, RUP = 0x0010;
}
"@
[System.Windows.Forms.SendKeys]::SendWait('{ESC}')
Start-Sleep -Milliseconds 300

# deplie le debordement
[SourisD]::SetCursorPos(1087, 697) | Out-Null
Start-Sleep -Milliseconds 300
[SourisD]::mouse_event([SourisD]::LDOWN, 0, 0, 0, [IntPtr]::Zero)
Start-Sleep -Milliseconds 60
[SourisD]::mouse_event([SourisD]::LUP, 0, 0, 0, [IntPtr]::Zero)
Start-Sleep -Milliseconds 900

# clic droit sur l'icone du singe dans le popup
[SourisD]::SetCursorPos(1048, 634) | Out-Null
Start-Sleep -Milliseconds 300
[SourisD]::mouse_event([SourisD]::RDOWN, 0, 0, 0, [IntPtr]::Zero)
Start-Sleep -Milliseconds 80
[SourisD]::mouse_event([SourisD]::RUP, 0, 0, 0, [IntPtr]::Zero)
Start-Sleep -Milliseconds 1500

$ecran = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds
$bmp = New-Object System.Drawing.Bitmap $ecran.Width, $ecran.Height
$g = [System.Drawing.Graphics]::FromImage($bmp)
$g.CopyFromScreen(0, 0, 0, 0, $ecran.Size)
$g.Dispose()
$bmp.Save('C:\Users\glance\SingeTest\menu.png')
$bmp.Dispose()
