package config

import "testing"

func TestDefaultsUsesUsernameEnvironmentVariable(t *testing.T) {
	t.Setenv("USERNAME", "fielduser")
	t.Setenv("USER", "unixuser")

	settings := Defaults()

	if settings.User != "FIELDUSER" {
		t.Fatalf("got user %q, want %q", settings.User, "FIELDUSER")
	}
}

func TestDefaultsFallsBackToUserEnvironmentVariable(t *testing.T) {
	t.Setenv("USERNAME", "")
	t.Setenv("USER", "unixuser")

	settings := Defaults()

	if settings.User != "UNIXUSER" {
		t.Fatalf("got user %q, want %q", settings.User, "UNIXUSER")
	}
}
