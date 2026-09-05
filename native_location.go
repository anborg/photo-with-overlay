//go:build !darwin

package main

import "fmt"

func getCurrentLocation() (CurrentLocation, error) {
	return CurrentLocation{}, fmt.Errorf("native location is unavailable on this platform")
}
