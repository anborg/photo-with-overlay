[CmdletBinding()]
param(
  [switch]$Browser,
  [string]$LogLevel = 'Debug'
)
$ErrorActionPreference = 'Stop'; Set-StrictMode -Version Latest; Push-Location $PSScriptRoot
try {
  # `wails dev` builds a debug binary with DevTools always enabled (right-click > Inspect,
  # or F12, works even though the release build from build.ps1 has DevTools off) and serves
  # the frontend from disk with hot reload instead of the embedded production bundle. That
  # also means no Cache-Control caching gotchas like the embedded release build can hit.
  #
  # -Browser additionally opens the app at its Vite dev server URL (default
  # http://localhost:5173) in your default system browser. Wails wires up a websocket bridge
  # so window.go/window.runtime calls still work there, so the real backend (camera, photo
  # storage, thumbnails, etc.) works in an ordinary browser tab too - handy for using
  # standard browser/WebView2 DevTools side by side, or comparing rendering against a
  # different browser engine while this gallery-thumbnail bug is being tracked down.
  $devArgs = @('-loglevel', $LogLevel)
  if ($Browser) { $devArgs += '-browser' }

  go run github.com/wailsapp/wails/v2/cmd/wails@v2.15.0 dev @devArgs
  if ($LASTEXITCODE -ne 0) { throw "wails dev exited with code $LASTEXITCODE." }
} finally {
  Pop-Location
}
