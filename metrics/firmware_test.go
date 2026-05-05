package metrics

import (
	"testing"

	"safely-you-homework/adapters"
)

func fwSample(version string) adapters.StoredSample {
	return adapters.StoredSample{Body: firmwareBody{Version: version}}
}

func TestFirmwareEmpty(t *testing.T) {
	def, ok := Lookup("firmware")
	if !ok {
		t.Fatal("firmware metric not registered")
	}
	if got := def.Aggregate(nil); got != nil {
		t.Errorf("expected nil for empty samples, got %v", got)
	}
}

func TestFirmwareLatestWins(t *testing.T) {
	def, _ := Lookup("firmware")
	samples := []adapters.StoredSample{
		fwSample("1.0.0"),
		fwSample("1.1.0"),
		fwSample("1.2.3"),
	}
	if got := def.Aggregate(samples); got != "1.2.3" {
		t.Errorf("expected latest version '1.2.3', got %v", got)
	}
}

func TestFirmwareIgnoresWrongType(t *testing.T) {
	def, _ := Lookup("firmware")
	samples := []adapters.StoredSample{
		fwSample("1.0.0"),
		{Body: "not a firmware body"},
	}
	if got := def.Aggregate(samples); got != "1.0.0" {
		t.Errorf("expected '1.0.0' (last sample wrong-type, skipped), got %v", got)
	}
}
