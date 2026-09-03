[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

$sourcePath = Join-Path $PSScriptRoot '20260903_103712_Prem_0001.jpg'
$buildIconPath = Join-Path $PSScriptRoot 'build\appicon.png'
$webIconPath = Join-Path $PSScriptRoot 'frontend\dist\appicon.png'
$source = [System.Drawing.Image]::FromFile($sourcePath)
$canvas = New-Object System.Drawing.Bitmap 1024, 1024
$graphics = [System.Drawing.Graphics]::FromImage($canvas)

try {
    $graphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
    $graphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
    # Use only the clean upper field of the source; exclude its personal metadata overlay.
    $side = [Math]::Min($source.Width, [Math]::Min($source.Height, 440))
    $sourceX = [int](($source.Width - $side) / 2)
    $sourceY = 0
    $graphics.DrawImage($source, [System.Drawing.Rectangle]::new(0, 0, 1024, 1024), $sourceX, $sourceY, $side, $side, [System.Drawing.GraphicsUnit]::Pixel)
    $graphics.FillRectangle([System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(105, 8, 12, 18)), 0, 0, 1024, 1024)
    $graphics.FillRectangle([System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(205, 18, 23, 31)), 118, 230, 788, 565)

    $whitePen = [System.Drawing.Pen]::new([System.Drawing.Color]::White, 48)
    $whitePen.LineJoin = [System.Drawing.Drawing2D.LineJoin]::Round
    $graphics.DrawRectangle($whitePen, 240, 345, 544, 336)
    $graphics.DrawEllipse($whitePen, 398, 397, 228, 228)
    $graphics.DrawLine($whitePen, 335, 345, 390, 285)
    $graphics.DrawLine($whitePen, 390, 285, 565, 285)
    $graphics.DrawLine($whitePen, 565, 285, 620, 345)

    $accentBrush = [System.Drawing.SolidBrush]::new([System.Drawing.Color]::FromArgb(232, 64, 56))
    $graphics.FillEllipse($accentBrush, 665, 565, 190, 190)
    $pinPen = [System.Drawing.Pen]::new([System.Drawing.Color]::White, 34)
    $graphics.DrawEllipse($pinPen, 712, 603, 96, 96)
    $graphics.DrawLine($pinPen, 760, 710, 760, 805)
    $graphics.FillRectangle($accentBrush, 180, 730, 415, 28)
    $graphics.FillRectangle([System.Drawing.Brushes]::White, 180, 775, 315, 22)

    New-Item -ItemType Directory -Force (Split-Path $buildIconPath), (Split-Path $webIconPath) | Out-Null
    $canvas.Save($buildIconPath, [System.Drawing.Imaging.ImageFormat]::Png)
    $canvas.Save($webIconPath, [System.Drawing.Imaging.ImageFormat]::Png)
}
finally {
    $graphics.Dispose()
    $canvas.Dispose()
    $source.Dispose()
}

Write-Host "Created app icons from $sourcePath" -ForegroundColor Green
