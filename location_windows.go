//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
)

func windowsLocation() (float64, float64, *float64, error) {
	const script = `
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$null = [Windows.Devices.Geolocation.Geolocator,Windows.Devices.Geolocation,ContentType=WindowsRuntime]
$locator = [Windows.Devices.Geolocation.Geolocator]::new()
$operation = $locator.GetGeopositionAsync()
$asTask = [System.WindowsRuntimeSystemExtensions].GetMethods() | Where-Object {
    $_.Name -eq 'AsTask' -and $_.IsGenericMethod -and $_.GetParameters().Count -eq 1
} | Select-Object -First 1
$task = $asTask.MakeGenericMethod([Windows.Devices.Geolocation.Geoposition]).Invoke($null, @($operation))
$task.Wait(20_000)
if (-not $task.IsCompletedSuccessfully) { throw 'Windows location request did not complete.' }
$position = $task.Result.Coordinate.Point.Position
[pscustomobject]@{ latitude=$position.Latitude; longitude=$position.Longitude; accuracy=$task.Result.Coordinate.Accuracy } | ConvertTo-Json -Compress
`
	output, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).CombinedOutput()
	if err != nil {
		return 0, 0, nil, fmt.Errorf("Windows location: %s", string(output))
	}
	var result struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Accuracy  float64 `json:"accuracy"`
	}
	if err = json.Unmarshal(output, &result); err != nil {
		return 0, 0, nil, fmt.Errorf("Windows location response: %w", err)
	}
	return result.Latitude, result.Longitude, &result.Accuracy, nil
}
