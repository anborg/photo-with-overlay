# Photo with GPS Overlay — Go

Maintainable Go/Wails implementation for Windows 11 and the Panasonic FZ-G2. WebView2 provides the touch UI and camera preview; the Go backend owns configuration, reverse geocoding, safe file operations, thumbnails, sequential filenames, and standards-compatible EXIF/XMP injection.

## Build

Requirements: Go 1.25+ and WebView2 (included with Windows 11). The frontend is intentionally static, so Node.js is not required for normal builds.

```powershell
.\run.ps1
.\build.ps1
```

Use `.\build.ps1 -SkipTests` only when local application-control policy prevents
newly compiled Go test executables from running.

The release executable is written to `build\bin\PhotoWithOverlay.exe`. Windows camera and location privacy permissions must be enabled. Settings are stored under the current user's configuration directory; photos default to `Pictures\Field Photos`.

The reverse-geocoded watermark line contains only street, city, province code, and postal code.
