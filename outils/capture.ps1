# Capture l'ecran complet de la session interactive.
# Doit tourner dans la session de l'utilisateur : depuis SSH (session 0) la
# capture serait noire, faute de bureau.
Add-Type -AssemblyName System.Windows.Forms, System.Drawing

$ecran = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds
$image = New-Object System.Drawing.Bitmap $ecran.Width, $ecran.Height
$g = [System.Drawing.Graphics]::FromImage($image)
$g.CopyFromScreen($ecran.Location, [System.Drawing.Point]::Empty, $ecran.Size)

$sortie = Join-Path $PSScriptRoot 'capture.png'
$image.Save($sortie, [System.Drawing.Imaging.ImageFormat]::Png)

$g.Dispose()
$image.Dispose()
