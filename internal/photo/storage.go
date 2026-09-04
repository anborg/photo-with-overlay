package photo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aamkam/photo-with-overlay/internal/jpegmeta"
)

// Save embeds capture metadata in JPEG data and writes it under the next
// available sequence number in folder.
func (s *Service) Save(jpegData []byte, meta Metadata, folder string) (Item, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(jpegData) < 2 || jpegData[0] != 0xff || jpegData[1] != 0xd8 {
		return Item{}, fmt.Errorf("captured data is not a JPEG")
	}
	if err := os.MkdirAll(folder, 0755); err != nil {
		return Item{}, fmt.Errorf("create output folder: %w", err)
	}

	prefix := fmt.Sprintf("%s_%s_", meta.CapturedAt.Format("20060102_150405"), safeUser(meta.User))
	sequence, err := nextSequence(folder, prefix)
	if err != nil {
		return Item{}, fmt.Errorf("determine next photo sequence: %w", err)
	}
	name := fmt.Sprintf("%s%04d.jpg", prefix, sequence)
	path := filepath.Join(folder, name)
	data, err := jpegmeta.Insert(jpegData, jpegmeta.Metadata{
		CapturedAt: meta.CapturedAt,
		User:       meta.User,
		Latitude:   meta.Latitude,
		Longitude:  meta.Longitude,
		Accuracy:   meta.Accuracy,
		Location:   meta.Location,
		Source:     meta.LocationSource,
	})
	if err != nil {
		return Item{}, fmt.Errorf("insert JPEG metadata: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return Item{}, fmt.Errorf("write photo %q: %w", path, err)
	}
	return Item{Name: name, Path: path, ModifiedAt: time.Now()}, nil
}

// List returns JPEG photos in folder ordered from newest to oldest.
// A folder that does not yet exist is treated as an empty gallery.
func (s *Service) List(folder string) ([]Item, error) {
	entries, err := os.ReadDir(folder)
	if os.IsNotExist(err) {
		return []Item{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read output folder: %w", err)
	}

	items := make([]Item, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".jpg") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("read photo information for %q: %w", entry.Name(), err)
		}
		items = append(items, Item{
			Name:       entry.Name(),
			Path:       filepath.Join(folder, entry.Name()),
			ModifiedAt: info.ModTime(),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ModifiedAt.After(items[j].ModifiedAt)
	})
	return items, nil
}

// ValidPhotoPath returns an absolute JPEG path when path is contained within
// folder. It rejects other extensions and lexical directory traversal.
func ValidPhotoPath(path, folder string) (string, error) {
	root, err := filepath.Abs(folder)
	if err != nil {
		return "", fmt.Errorf("resolve output folder: %w", err)
	}
	clean, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve photo path: %w", err)
	}
	relative, err := filepath.Rel(root, clean)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("photo is outside the output folder")
	}
	if !strings.EqualFold(filepath.Ext(clean), ".jpg") {
		return "", fmt.Errorf("only JPEG photos may be accessed")
	}
	return clean, nil
}

func safeUser(user string) string {
	var safe strings.Builder
	for _, character := range user {
		if character >= 'a' && character <= 'z' ||
			character >= 'A' && character <= 'Z' ||
			character >= '0' && character <= '9' ||
			character == '-' || character == '_' {
			safe.WriteRune(character)
		}
	}
	if safe.Len() == 0 {
		return "USER"
	}
	return safe.String()
}

func nextSequence(folder, prefix string) (int, error) {
	matches, err := filepath.Glob(filepath.Join(folder, prefix+"*.jpg"))
	if err != nil {
		return 0, err
	}
	maximum := 0
	for _, path := range matches {
		var sequence int
		name := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(path), prefix), ".jpg")
		if _, err := fmt.Sscanf(name, "%d", &sequence); err == nil && sequence > maximum {
			maximum = sequence
		}
	}
	return maximum + 1, nil
}
