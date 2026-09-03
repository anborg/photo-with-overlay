package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"time"

	"github.com/aamkam/photo-with-overlay/internal/config"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

type captureLocation struct {
	latitude, longitude float64
	accuracy            *float64
	address, roadClue   string
	source              string
}

type overlayLine struct {
	text string
	size float64
}

func drawOverlay(source image.Image, settings config.Settings, when time.Time, location captureLocation) (image.Image, error) {
	bounds := source.Bounds()
	destination := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	draw.Draw(destination, destination.Bounds(), source, bounds.Min, draw.Src)
	mainSize := float64(max(12, min(120, settings.FontSize)))
	lines := []overlayLine{
		{when.Format("2006-01-02 15:04:05"), mainSize},
		{"User: " + settings.User, mainSize},
		{fmt.Sprintf("%.6f, %.6f", location.latitude, location.longitude), mainSize},
	}
	if location.address != "" {
		lines = append(lines, overlayLine{location.address, mainSize})
	}
	if location.roadClue != "" {
		lines = append(lines, overlayLine{location.roadClue, math.Max(10, math.Round(mainSize*.62))})
	}
	parsedFont, err := opentype.Parse(gobold.TTF)
	if err != nil {
		return nil, err
	}
	type measuredLine struct {
		overlayLine
		face   font.Face
		width  int
		height int
	}
	measured := make([]measuredLine, 0, len(lines))
	maxWidth, totalHeight := 0, 0
	const padding, gap = 20, 6
	for _, line := range lines {
		face, faceErr := opentype.NewFace(parsedFont, &opentype.FaceOptions{Size: line.size, DPI: 72, Hinting: font.HintingFull})
		if faceErr != nil {
			return nil, faceErr
		}
		width := font.MeasureString(face, line.text).Ceil()
		height := int(math.Ceil(line.size * 1.2))
		measured = append(measured, measuredLine{overlayLine: line, face: face, width: width, height: height})
		maxWidth = max(maxWidth, width)
		totalHeight += height
	}
	panelWidth := maxWidth + padding*2
	panelHeight := totalHeight + gap*(len(lines)-1) + padding*2
	x, y := padding, padding
	if len(settings.WatermarkPosition) >= 5 && settings.WatermarkPosition[len(settings.WatermarkPosition)-5:] == "right" {
		x = destination.Bounds().Dx() - panelWidth - padding
	}
	if len(settings.WatermarkPosition) >= 6 && settings.WatermarkPosition[:6] == "bottom" {
		y = destination.Bounds().Dy() - panelHeight - padding
	}
	x, y = max(0, x), max(0, y)
	draw.Draw(destination, image.Rect(x, y, x+panelWidth, y+panelHeight), &image.Uniform{C: color.RGBA{0, 0, 0, 172}}, image.Point{}, draw.Over)
	lineY := y + padding
	for _, line := range measured {
		drawer := &font.Drawer{Dst: destination, Src: image.White, Face: line.face, Dot: fixedPoint(x+padding, lineY+line.face.Metrics().Ascent.Ceil())}
		drawer.DrawString(line.text)
		lineY += line.height + gap
		_ = line.face.Close()
	}
	return destination, nil
}

func fixedPoint(x, y int) fixed.Point26_6 { return fixed.P(x, y) }
