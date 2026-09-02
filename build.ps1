[CmdletBinding()]param([switch]$SkipTests)
$ErrorActionPreference='Stop';Set-StrictMode -Version Latest;Push-Location $PSScriptRoot
try {
  if(-not $SkipTests){go test ./...;if($LASTEXITCODE-ne 0){throw "Tests failed ($LASTEXITCODE)."}}
  New-Item -ItemType Directory -Force '.\build\bin'|Out-Null
  go build -tags "desktop,production" -trimpath -ldflags "-s -w -H windowsgui" -o '.\build\bin\PhotoWithOverlay.exe' .
  if($LASTEXITCODE-ne 0){throw "Build failed ($LASTEXITCODE)."}
  $exe=Get-Item '.\build\bin\PhotoWithOverlay.exe';Write-Host "Built: $($exe.FullName) ($([math]::Round($exe.Length/1MB,1)) MB)" -ForegroundColor Green
}finally{Pop-Location}
