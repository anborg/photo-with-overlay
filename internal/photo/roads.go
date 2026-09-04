package photo

import (
	"encoding/json"
	"io"
	"math"
	"sort"
	"strings"
)

type roadCandidate struct {
	name     string
	distance float64
}

// ParseNearbyRoads returns the closest distinct named roads from an Overpass
// response, excluding the road that contains the current location.
func ParseNearbyRoads(reader io.Reader, currentRoad string, latitude, longitude float64, limit int) ([]string, error) {
	var payload struct {
		Elements []struct {
			Tags     map[string]string `json:"tags"`
			Geometry []struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"geometry"`
		} `json:"elements"`
	}
	if err := json.NewDecoder(reader).Decode(&payload); err != nil {
		return nil, err
	}

	closestByName := make(map[string]float64)
	for _, element := range payload.Elements {
		name := strings.TrimSpace(element.Tags["name"])
		if name == "" || strings.EqualFold(name, strings.TrimSpace(currentRoad)) {
			continue
		}
		distance := math.Inf(1)
		for _, point := range element.Geometry {
			distance = math.Min(distance, haversineMetres(latitude, longitude, point.Lat, point.Lon))
		}
		if previous, exists := closestByName[name]; !exists || distance < previous {
			closestByName[name] = distance
		}
	}

	candidates := make([]roadCandidate, 0, len(closestByName))
	for name, distance := range closestByName {
		candidates = append(candidates, roadCandidate{name: name, distance: distance})
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].distance < candidates[j].distance
	})
	if limit < 0 || limit > len(candidates) {
		limit = len(candidates)
	}
	roads := make([]string, limit)
	for index := range roads {
		roads[index] = candidates[index].name
	}
	return roads, nil
}

func haversineMetres(latitude1, longitude1, latitude2, longitude2 float64) float64 {
	const earthRadiusMetres = 6371000
	const radiansPerDegree = math.Pi / 180
	deltaLatitude := (latitude2 - latitude1) * radiansPerDegree
	deltaLongitude := (longitude2 - longitude1) * radiansPerDegree
	a := math.Sin(deltaLatitude/2)*math.Sin(deltaLatitude/2) +
		math.Cos(latitude1*radiansPerDegree)*math.Cos(latitude2*radiansPerDegree)*
			math.Sin(deltaLongitude/2)*math.Sin(deltaLongitude/2)
	return earthRadiusMetres * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}
