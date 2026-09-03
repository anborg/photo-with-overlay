package jpegmeta

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
	"time"
)

func TestInsertAddsExifAndXMPAndPreservesJPEG(t *testing.T) {
	var source bytes.Buffer
	sourceImage := image.NewRGBA(image.Rect(0, 0, 1, 1))
	sourceImage.Set(0, 0, color.RGBA{20, 40, 60, 255})
	if err := jpeg.Encode(&source, sourceImage, nil); err != nil {
		t.Fatal(err)
	}
	accuracy := 4.5
	when := time.Date(2026, 9, 2, 15, 31, 42, 0, time.FixedZone("EDT", -4*60*60))
	result, err := Insert(source.Bytes(), Metadata{CapturedAt: when, User: "JSMITH", Latitude: 43.8561, Longitude: -79.337, Accuracy: &accuracy, Location: "Main Street, Markham, ON, L3P 1A1", Source: "Windows Location"})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(result, []byte("Exif\x00\x00")) {
		t.Error("EXIF APP1 missing")
	}
	if !bytes.Contains(result, []byte("http://ns.adobe.com/xap/1.0/\x00")) {
		t.Error("XMP APP1 missing")
	}
	if !bytes.Contains(result, []byte("JSMITH")) {
		t.Error("operator missing")
	}
	if _, err = jpeg.Decode(bytes.NewReader(result)); err != nil {
		t.Fatalf("output is not a valid JPEG: %v", err)
	}
}
