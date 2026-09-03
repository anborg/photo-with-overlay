[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$executablePath = Join-Path $PSScriptRoot 'build\bin\PhotoWithOverlay.exe'

if (-not (Test-Path -LiteralPath $executablePath -PathType Leaf)) {
    throw "Application not found at '$executablePath'. Run .\buildrelease.ps1 first."
}

& $executablePath
