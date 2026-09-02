package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type Settings struct {
	User              string  `json:"user"`
	OutputFolder      string  `json:"outputFolder"`
	WatermarkPosition string  `json:"watermarkPosition"`
	FontFamily        string  `json:"fontFamily"`
	FontSize          int     `json:"fontSize"`
	UseManualLocation bool    `json:"useManualLocation"`
	ManualLatitude    float64 `json:"manualLatitude"`
	ManualLongitude   float64 `json:"manualLongitude"`
	ManualAddress     string  `json:"manualAddress"`
	ReverseGeocode    bool    `json:"reverseGeocode"`
	CameraID          string  `json:"cameraId"`
}

func Defaults() Settings {
	home, _ := os.UserHomeDir()
	return Settings{User: strings.ToUpper(os.Getenv("USERNAME")), OutputFolder: filepath.Join(home, "Pictures", "Field Photos"),
		WatermarkPosition: "bottom-left", FontFamily: "Arial", FontSize: 36, UseManualLocation: true,
		ManualLatitude: 43.8561, ManualLongitude: -79.3370, ManualAddress: "Markham, ON"}
}

func path() string {
	dir, _ := os.UserConfigDir()
	return filepath.Join(dir, "PhotoWithOverlayGo", "settings.json")
}
func Load() Settings {
	s := Defaults()
	data, err := os.ReadFile(path())
	if err == nil {
		_ = json.Unmarshal(data, &s)
	}
	return s
}
func Save(s Settings) error {
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

type ValidationError struct{ Message string }

func (e *ValidationError) Error() string { return e.Message }
