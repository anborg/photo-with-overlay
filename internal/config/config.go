// Package config persists user-selectable application settings.
package config

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// Settings contains frontend, camera, output, and location preferences.
type Settings struct {
	User              string  `json:"user"`
	OutputFolder      string  `json:"outputFolder"`
	WatermarkPosition string  `json:"watermarkPosition"`
	WatermarkX        float64 `json:"watermarkX"`
	WatermarkY        float64 `json:"watermarkY"`
	WatermarkWidth    float64 `json:"watermarkWidth"`
	FontFamily        string  `json:"fontFamily"`
	FontSize          int     `json:"fontSize"`
	UseManualLocation bool    `json:"useManualLocation"`
	ManualLatitude    float64 `json:"manualLatitude"`
	ManualLongitude   float64 `json:"manualLongitude"`
	ManualAddress     string  `json:"manualAddress"`
	ReverseGeocode    bool    `json:"reverseGeocode"`
	CameraID          string  `json:"cameraId"`
}

// Defaults returns settings suitable for a first application launch.
func Defaults() Settings {
	home, _ := os.UserHomeDir()
	return Settings{User: defaultUser(), OutputFolder: filepath.Join(home, "Pictures", "Field Photos"),
		WatermarkPosition: "bottom-left", WatermarkX: 0, WatermarkY: 1, WatermarkWidth: 0.42, FontFamily: "Arial", FontSize: 36, UseManualLocation: true,
		ManualLatitude: 43.8561, ManualLongitude: -79.3370, ManualAddress: "Markham, ON"}
}

func defaultUser() string {
	for _, candidate := range []string{os.Getenv("USERNAME"), os.Getenv("USER")} {
		if trimmed := strings.TrimSpace(candidate); trimmed != "" {
			return strings.ToUpper(trimmed)
		}
	}
	if current, err := user.Current(); err == nil {
		for _, candidate := range []string{current.Username, current.Name} {
			if trimmed := strings.TrimSpace(candidate); trimmed != "" {
				return strings.ToUpper(trimmed)
			}
		}
	}
	return "OPERATOR"
}

func path() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "PhotoWithOverlayGo", "settings.json")
}

// Load reads persisted settings, falling back to defaults when the settings
// file is absent or cannot be decoded.
func Load() Settings {
	s := Defaults()
	data, err := os.ReadFile(path())
	if err == nil {
		_ = json.Unmarshal(data, &s)
	}
	if strings.TrimSpace(s.User) == "" {
		s.User = defaultUser()
	}
	return s
}

// Save validates and serializes settings to the user's config directory.
func Save(s Settings) error {
	if strings.TrimSpace(s.User) == "" {
		s.User = defaultUser()
	}
	if strings.TrimSpace(s.User) == "" {
		return &ValidationError{"operator is required"}
	}
	if strings.TrimSpace(s.OutputFolder) == "" {
		return &ValidationError{"output folder is required"}
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path()), 0755); err != nil {
		return err
	}
	return os.WriteFile(path(), data, 0600)
}

// ValidationError reports a user-correctable settings problem.
type ValidationError struct{ Message string }

// Error implements the error interface.
func (e *ValidationError) Error() string { return e.Message }
