package photo

import (
	"strings"
	"testing"
)

func TestParseNearbyRoadsReturnsClosestDistinctRoads(t *testing.T) {
	input := `{"elements":[
		{"tags":{"name":"Highway 7"},"geometry":[{"lat":43.8565,"lon":-79.3373}]},
		{"tags":{"name":"Warden Avenue"},"geometry":[{"lat":43.8566,"lon":-79.3374}]},
		{"tags":{"name":"Village Parkway"},"geometry":[{"lat":43.8580,"lon":-79.3400}]},
		{"tags":{"name":"Warden Avenue"},"geometry":[{"lat":43.8600,"lon":-79.3500}]}
	]}`
	roads, err := ParseNearbyRoads(strings.NewReader(input), "Highway 7", 43.856577, -79.337355, 2)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := strings.Join(roads, " / "), "Warden Avenue / Village Parkway"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
