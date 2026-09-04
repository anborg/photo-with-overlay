package photo

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"os"

	"golang.org/x/image/draw"
)

// Thumbnail scales a photo to 280 pixels wide while preserving its aspect
// ratio, then returns it as a JPEG data URL for display by the frontend.
func (s *Service) Thumbnail(path, folder string) (string, error) {
	cleanPath, err := ValidPhotoPath(path, folder)
	if err != nil {
		return "", err
	}
	file, err := os.Open(cleanPath)
	if err != nil {
		return "", fmt.Errorf("open photo for thumbnail: %w", err)
	}
	defer file.Close()

	source, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("decode thumbnail source: %w", err)
	}
	if source.Bounds().Dx() <= 0 || source.Bounds().Dy() <= 0 {
		return "", fmt.Errorf("photo has invalid dimensions")
	}

	const width = 280
	height := max(1, source.Bounds().Dy()*width/source.Bounds().Dx())
	destination := image.NewRGBA(image.Rect(0, 0, width, height))
	// ApproxBiLinear provides a useful quality/speed balance for small gallery previews.
	draw.ApproxBiLinear.Scale(destination, destination.Bounds(), source, source.Bounds(), draw.Over, nil)

	var output bytes.Buffer
	if err := jpeg.Encode(&output, destination, &jpeg.Options{Quality: 72}); err != nil {
		return "", fmt.Errorf("encode thumbnail: %w", err)
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(output.Bytes()), nil
}
