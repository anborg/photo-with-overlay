package photo

import (
	"strings"
	"testing"
)

func TestParseNominatimAddressOmitsRegions(t *testing.T) {
	input := `{"address":{"road":"Redfinch Crescent","city":"Vaughan","county":"York Region","region":"Golden Horseshoe","state":"Ontario","ISO3166-2-lvl4":"CA-ON","postcode":"L6A 4B2","country":"Canada"}}`
	got, err := ParseNominatimAddress(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if want := "Redfinch Crescent, Vaughan, ON, L6A 4B2"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestValidPhotoPathRejectsEscape(t *testing.T) {
	if _, err := ValidPhotoPath(`C:\outside\photo.jpg`, `C:\photos`); err == nil {
		t.Fatal("expected path escape to be rejected")
	}
}
