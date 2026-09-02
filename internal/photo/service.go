package photo

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aamkam/photo-with-overlay/internal/jpegmeta"
	"golang.org/x/image/draw"
)

type Service struct{ mu sync.Mutex }

func NewService() *Service { return &Service{} }

type Metadata struct {
	CapturedAt               time.Time
	User                     string
	Latitude, Longitude      float64
	Accuracy                 *float64
	Location, LocationSource string
}
type Item struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	ModifiedAt time.Time `json:"modifiedAt"`
}

func (s *Service) Save(jpegData []byte, meta Metadata, folder string) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(jpegData) < 2 || jpegData[0] != 0xff || jpegData[1] != 0xd8 {
		return Item{}, fmt.Errorf("captured data is not a JPEG")
	}
	if err := os.MkdirAll(folder, 0755); err != nil {
		return Item{}, err
	}
	user := safeUser(meta.User)
	prefix := fmt.Sprintf("%s_%s_", meta.CapturedAt.Format("20060102_150405"), user)
	sequence, err := nextSequence(folder, prefix)
	if err != nil {
		return Item{}, err
	}
	name := fmt.Sprintf("%s%04d.jpg", prefix, sequence)
	path := filepath.Join(folder, name)
	data, err := jpegmeta.Insert(jpegData, jpegmeta.Metadata{CapturedAt: meta.CapturedAt, User: meta.User,
		Latitude: meta.Latitude, Longitude: meta.Longitude, Accuracy: meta.Accuracy, Location: meta.Location, Source: meta.LocationSource})
	if err != nil {
		return Item{}, err
	}
	if err = os.WriteFile(path, data, 0644); err != nil {
		return Item{}, err
	}
	return Item{Name: name, Path: path, ModifiedAt: time.Now()}, nil
}

func (s *Service) List(folder string) ([]Item, error) {
	entries, err := os.ReadDir(folder)
	if os.IsNotExist(err) {
		return []Item{}, nil
	}
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jpg") {
			continue
		}
		info, e := entry.Info()
		if e == nil {
			items = append(items, Item{Name: entry.Name(), Path: filepath.Join(folder, entry.Name()), ModifiedAt: info.ModTime()})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ModifiedAt.After(items[j].ModifiedAt) })
	return items, nil
}

func (s *Service) Thumbnail(path, folder string) (string, error) {
	clean, err := ValidPhotoPath(path, folder)
	if err != nil {
		return "", err
	}
	f, err := os.Open(clean)
	if err != nil {
		return "", err
	}
	defer f.Close()
	src, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}
	w := 280
	h := src.Bounds().Dy() * w / src.Bounds().Dx()
	if h < 1 {
		h = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	var out bytes.Buffer
	if err = jpeg.Encode(&out, dst, &jpeg.Options{Quality: 72}); err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(out.Bytes()), nil
}

func ValidPhotoPath(path, folder string) (string, error) {
	root, err := filepath.Abs(folder)
	if err != nil {
		return "", err
	}
	clean, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, clean)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("photo is outside the output folder")
	}
	if !strings.EqualFold(filepath.Ext(clean), ".jpg") {
		return "", fmt.Errorf("only JPEG photos may be accessed")
	}
	return clean, nil
}

func ParseNominatimAddress(r io.Reader) (string, error) {
	var payload struct {
		Address map[string]string `json:"address"`
	}
	if err := json.NewDecoder(r).Decode(&payload); err != nil {
		return "", err
	}
	street := first(payload.Address, "road", "pedestrian", "footway", "path")
	city := first(payload.Address, "city", "town", "village", "municipality")
	province := provinceCode(payload.Address)
	postal := first(payload.Address, "postcode")
	parts := []string{}
	for _, value := range []string{street, city, province, postal} {
		if strings.TrimSpace(value) != "" {
			parts = append(parts, strings.TrimSpace(value))
		}
	}
	return strings.Join(parts, ", "), nil
}
func first(values map[string]string, keys ...string) string {
	for _, k := range keys {
		if values[k] != "" {
			return values[k]
		}
	}
	return ""
}
func provinceCode(a map[string]string) string {
	if iso := a["ISO3166-2-lvl4"]; strings.Contains(iso, "-") {
		p := strings.Split(iso, "-")
		return strings.ToUpper(p[len(p)-1])
	}
	provinces := map[string]string{"Ontario": "ON", "Quebec": "QC", "Alberta": "AB", "British Columbia": "BC", "Manitoba": "MB", "New Brunswick": "NB", "Newfoundland and Labrador": "NL", "Nova Scotia": "NS", "Prince Edward Island": "PE", "Saskatchewan": "SK", "Northwest Territories": "NT", "Nunavut": "NU", "Yukon": "YT"}
	if code := provinces[a["state"]]; code != "" {
		return code
	}
	return a["state"]
}
func safeUser(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "USER"
	}
	return b.String()
}
func nextSequence(folder, prefix string) (int, error) {
	matches, err := filepath.Glob(filepath.Join(folder, prefix+"*.jpg"))
	if err != nil {
		return 0, err
	}
	max := 0
	for _, p := range matches {
		var n int
		_, _ = fmt.Sscanf(strings.TrimSuffix(strings.TrimPrefix(filepath.Base(p), prefix), ".jpg"), "%d", &n)
		if n > max {
			max = n
		}
	}
	return max + 1, nil
}
