[CmdletBinding(SupportsShouldProcess)]
param()

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$projectRoot = [System.IO.Path]::GetFullPath($PSScriptRoot)
$buildOutput = [System.IO.Path]::GetFullPath((Join-Path $projectRoot 'build\bin'))
$expectedOutput = $projectRoot.TrimEnd([System.IO.Path]::DirectorySeparatorChar) +
    [System.IO.Path]::DirectorySeparatorChar + 'build' +
    [System.IO.Path]::DirectorySeparatorChar + 'bin'

if ($buildOutput -ne $expectedOutput) {
    throw "Refusing to clean unexpected path: $buildOutput"
}

if (-not (Test-Path -LiteralPath $buildOutput)) {
    Write-Host 'Already clean: build\bin does not exist.'
    return
}

if ($PSCmdlet.ShouldProcess($buildOutput, 'Remove build output')) {
    Remove-Item -LiteralPath $buildOutput -Recurse -Force
    Write-Host "Cleaned: $buildOutput" -ForegroundColor Green
}
