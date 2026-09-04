[CmdletBinding()]param([switch]$SkipTests)
$ErrorActionPreference='Stop';Set-StrictMode -Version Latest;Push-Location $PSScriptRoot
try {
  if(-not $SkipTests){go test ./...;if($LASTEXITCODE-ne 0){throw "Tests failed ($LASTEXITCODE)."}}
  go run github.com/wailsapp/wails/v2/cmd/wails@v2.11.0 build -s -skipbindings -trimpath -ldflags "-s -w" -o PhotoWithOverlay.exe
  if($LASTEXITCODE-ne 0){throw "Build failed ($LASTEXITCODE)."}
  $exe=Get-Item '.\build\bin\PhotoWithOverlay.exe';Write-Host "Built: $($exe.FullName) ($([math]::Round($exe.Length/1MB,1)) MB)" -ForegroundColor Green
}finally{Pop-Location}
