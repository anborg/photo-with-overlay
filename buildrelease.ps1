[CmdletBinding()]
param(
    [switch]$SkipTests,
    [switch]$SkipCompression
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
Push-Location $PSScriptRoot

try {
    if (-not $SkipTests) {
        go test ./...
        if ($LASTEXITCODE -ne 0) {
            throw "Tests failed ($LASTEXITCODE)."
        }
    }

    $outputDirectory = '.\build\bin'
    $outputPath = Join-Path $outputDirectory 'PhotoWithOverlay.exe'
    New-Item -ItemType Directory -Force $outputDirectory | Out-Null

    go run github.com/wailsapp/wails/v2/cmd/wails@v2.11.0 build `
        -s `
        -skipbindings `
        -trimpath `
        -ldflags '-s -w -buildid=' `
        -o PhotoWithOverlay.exe

    if ($LASTEXITCODE -ne 0) {
        throw "Release build failed ($LASTEXITCODE)."
    }

    $uncompressedSize = (Get-Item -LiteralPath $outputPath).Length
    $upx = Get-Command 'upx' -ErrorAction SilentlyContinue

    if (-not $SkipCompression -and $null -ne $upx) {
        $backupPath = "$outputPath.uncompressed"
        Copy-Item -LiteralPath $outputPath -Destination $backupPath -Force
        try {
            & $upx.Source --best --lzma --no-progress $outputPath
            if ($LASTEXITCODE -ne 0) {
                throw "UPX exited with code $LASTEXITCODE."
            }
        }
        catch {
            Copy-Item -LiteralPath $backupPath -Destination $outputPath -Force
            Write-Warning "UPX compression was unavailable; keeping the stripped binary. $($_.Exception.Message)"
        }
        finally {
            Remove-Item -LiteralPath $backupPath -Force -ErrorAction SilentlyContinue
        }
    }
    elseif (-not $SkipCompression) {
        Write-Warning 'UPX was not found; the binary is stripped but not executable-compressed.'
    }

    $exe = Get-Item -LiteralPath $outputPath
    $saved = $uncompressedSize - $exe.Length
    $compression = if ($uncompressedSize -gt 0) {
        [math]::Round(($saved / $uncompressedSize) * 100, 1)
    }
    else {
        0
    }

    Write-Host "Built release: $($exe.FullName)" -ForegroundColor Green
    Write-Host "Size: $([math]::Round($exe.Length / 1MB, 2)) MB (saved $compression%)" -ForegroundColor Green
}
finally {
    Pop-Location
}
