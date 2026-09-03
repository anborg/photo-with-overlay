package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aamkam/photo-with-overlay/internal/config"
	"github.com/aamkam/photo-with-overlay/internal/photo"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	photos *photo.Service
	client *http.Client
}

type SaveRequest struct {
	JPEGDataURL    string   `json:"jpegDataUrl"`
	CapturedAt     string   `json:"capturedAt"`
	User           string   `json:"user"`
	Latitude       float64  `json:"latitude"`
	Longitude      float64  `json:"longitude"`
	Accuracy       *float64 `json:"accuracy"`
	Location       string   `json:"location"`
	LocationSource string   `json:"locationSource"`
	OutputFolder   string   `json:"outputFolder"`
}

func NewApp() *App {
	return &App{photos: photo.NewService(), client: &http.Client{Timeout: 8 * time.Second}}
}
func (a *App) startup(ctx context.Context) { a.ctx = ctx }
func (a *App) shutdown(context.Context)    {}

func (a *App) LoadSettings() config.Settings               { return config.Load() }
func (a *App) SaveSettings(settings config.Settings) error { return config.Save(settings) }

func (a *App) SelectOutputFolder(current string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{Title: "Select photo output folder", DefaultDirectory: current})
}

func (a *App) SavePhoto(req SaveRequest) (photo.Item, error) {
	comma := strings.IndexByte(req.JPEGDataURL, ',')
	if comma < 0 {
		return photo.Item{}, fmt.Errorf("invalid JPEG data")
	}
	data, err := base64.StdEncoding.DecodeString(req.JPEGDataURL[comma+1:])
	if err != nil {
		return photo.Item{}, fmt.Errorf("decode JPEG: %w", err)
	}
	when, err := time.Parse(time.RFC3339Nano, req.CapturedAt)
	if err != nil {
		return photo.Item{}, fmt.Errorf("capture time: %w", err)
	}
	return a.photos.Save(data, photo.Metadata{CapturedAt: when, User: req.User, Latitude: req.Latitude,
		Longitude: req.Longitude, Accuracy: req.Accuracy, Location: req.Location, LocationSource: req.LocationSource}, req.OutputFolder)
}

func (a *App) ListPhotos(folder string) ([]photo.Item, error) { return a.photos.List(folder) }
func (a *App) Thumbnail(path, folder string) (string, error)  { return a.photos.Thumbnail(path, folder) }

func (a *App) ShowPhoto(path, folder string) error {
	clean, err := photo.ValidPhotoPath(path, folder)
	if err != nil {
		return err
	}
	return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", clean).Start()
}

func (a *App) DeletePhoto(path, folder string) error {
	clean, err := photo.ValidPhotoPath(path, folder)
	if err != nil {
		return err
	}
	return os.Remove(clean)
}

type LocationDetails struct {
	Address  string `json:"address"`
	RoadClue string `json:"roadClue"`
}

func (a *App) ReverseGeocode(latitude, longitude float64) (LocationDetails, error) {
	u := fmt.Sprintf("https://nominatim.openstreetmap.org/reverse?format=jsonv2&lat=%f&lon=%f&zoom=18", latitude, longitude)
	req, _ := http.NewRequestWithContext(a.ctx, http.MethodGet, u, nil)
	req.Header.Set("User-Agent", "PhotoWithOverlayGo/1.0 (municipal field-photo application)")
	resp, err := a.client.Do(req)
	if err != nil {
		return LocationDetails{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return LocationDetails{}, fmt.Errorf("reverse geocoder returned %s", resp.Status)
	}
	address, road, err := photo.ParseNominatimLocation(resp.Body)
	if err != nil {
		return LocationDetails{}, err
	}
	details := LocationDetails{Address: address}

	query := fmt.Sprintf(`[out:json][timeout:8];way(around:600,%f,%f)["highway"]["name"];out tags geom;`, latitude, longitude)
	form := url.Values{"data": {query}}
	overpassReq, _ := http.NewRequestWithContext(a.ctx, http.MethodPost, "https://overpass-api.de/api/interpreter", strings.NewReader(form.Encode()))
	overpassReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	overpassReq.Header.Set("User-Agent", "PhotoWithOverlayGo/1.0 (municipal field-photo application)")
	if overpassResp, requestErr := a.client.Do(overpassReq); requestErr == nil {
		defer overpassResp.Body.Close()
		if overpassResp.StatusCode == http.StatusOK {
			if roads, parseErr := photo.ParseNearbyRoads(overpassResp.Body, road, latitude, longitude, 2); parseErr == nil && len(roads) > 0 {
				details.RoadClue = "Nearby cross roads: " + strings.Join(roads, " / ")
			}
		}
	}
	return details, nil
}
