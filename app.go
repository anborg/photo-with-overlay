package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/aamkam/photo-with-overlay/internal/config"
	"github.com/aamkam/photo-with-overlay/internal/photo"
)

type App struct {
	ctx    context.Context
	photos *photo.Service
	client *http.Client
}

func NewApp() *App {
	return &App{ctx: context.Background(), photos: photo.NewService(), client: &http.Client{Timeout: 8 * time.Second}}
}

func (a *App) LoadSettings() config.Settings                  { return config.Load() }
func (a *App) SaveSettings(settings config.Settings) error    { return config.Save(settings) }
func (a *App) ListPhotos(folder string) ([]photo.Item, error) { return a.photos.List(folder) }

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
	Address  string
	RoadClue string
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
