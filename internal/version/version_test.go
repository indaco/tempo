package version //nolint:revive // package name matches the package under test

import (
	"testing"
)

// TestGetVersion checks if the GetVersion function correctly retrieves the embedded version.
func TestGetVersion(t *testing.T) {
	expectedVersion := "0.2.3"

	got := GetVersion()
	if got != expectedVersion {
		t.Errorf("GetVersion() = %q; want %q", got, expectedVersion)
	}
}
