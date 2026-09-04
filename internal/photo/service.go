// Package photo stores captured photos and prepares their location metadata.
package photo

import (
	"sync"
	"time"
)

// Service serializes photo writes so generated sequence numbers remain unique.
type Service struct {
	mu sync.Mutex
}

// NewService creates a photo service with serialized write access.
func NewService() *Service {
	return &Service{}
}

// Metadata describes the capture and location information embedded in a photo.
type Metadata struct {
	CapturedAt     time.Time
	User           string
	Latitude       float64
	Longitude      float64
	Accuracy       *float64
	Location       string
	LocationSource string
}

// Item describes a saved photo exposed to the frontend gallery.
type Item struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	ModifiedAt time.Time `json:"modifiedAt"`
}
