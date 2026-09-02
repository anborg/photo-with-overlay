[CmdletBinding()]param()
$ErrorActionPreference='Stop';Push-Location $PSScriptRoot
try{go run -tags "desktop,production" .;if($LASTEXITCODE-ne 0){throw "Application exited ($LASTEXITCODE)."}}finally{Pop-Location}
