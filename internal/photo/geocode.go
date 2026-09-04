package photo

import (
	"encoding/json"
	"io"
	"strings"
)

// ParseNominatimAddress extracts a compact Canadian mailing address from a
// Nominatim reverse-geocoding response.
func ParseNominatimAddress(reader io.Reader) (string, error) {
	address, _, err := ParseNominatimLocation(reader)
	return address, err
}

// ParseNominatimLocation extracts a compact address and its street name from a
// Nominatim reverse-geocoding response.
func ParseNominatimLocation(reader io.Reader) (string, string, error) {
	var payload struct {
		Address map[string]string `json:"address"`
	}
	if err := json.NewDecoder(reader).Decode(&payload); err != nil {
		return "", "", err
	}

	street := firstValue(payload.Address, "road", "pedestrian", "footway", "path")
	parts := []string{
		street,
		firstValue(payload.Address, "city", "town", "village", "municipality"),
		provinceCode(payload.Address),
		firstValue(payload.Address, "postcode"),
	}
	addressParts := make([]string, 0, len(parts))
	for _, part := range parts {
		if part = strings.TrimSpace(part); part != "" {
			addressParts = append(addressParts, part)
		}
	}
	return strings.Join(addressParts, ", "), street, nil
}

func firstValue(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if values[key] != "" {
			return values[key]
		}
	}
	return ""
}

func provinceCode(address map[string]string) string {
	if isoCode := address["ISO3166-2-lvl4"]; strings.Contains(isoCode, "-") {
		parts := strings.Split(isoCode, "-")
		return strings.ToUpper(parts[len(parts)-1])
	}
	provinces := map[string]string{
		"Alberta": "AB", "British Columbia": "BC", "Manitoba": "MB",
		"New Brunswick": "NB", "Newfoundland and Labrador": "NL", "Nova Scotia": "NS",
		"Northwest Territories": "NT", "Nunavut": "NU", "Ontario": "ON",
		"Prince Edward Island": "PE", "Quebec": "QC", "Saskatchewan": "SK", "Yukon": "YT",
	}
	if code := provinces[address["state"]]; code != "" {
		return code
	}
	return address["state"]
}
