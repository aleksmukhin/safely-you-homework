package metrics

import (
	"testing"
	"time"

	"safely-you-homework/adapters"
)

func utSample(uploadNs int) adapters.StoredSample {
	return adapters.StoredSample{IngestedAt: time.Now(), Body: uploadTimeBody{UploadTime: uploadNs}}
}

func TestAggregateUploadTimeEmpty(t *testing.T) {
	if got := aggregateUploadTime(nil).(string); got != "0s" {
		t.Errorf("expected '0s' for empty samples, got %v", got)
	}
}

func TestAggregateUploadTimeAverage(t *testing.T) {
	samples := []adapters.StoredSample{
		utSample(1_000_000_000),
		utSample(2_000_000_000),
		utSample(3_000_000_000),
	}
	if got := aggregateUploadTime(samples).(string); got != "2s" {
		t.Errorf("expected '2s' (avg of 1s/2s/3s), got %v", got)
	}
}

func TestAggregateUploadTimeIgnoresWrongType(t *testing.T) {
	samples := []adapters.StoredSample{
		utSample(2_000_000_000),
		{Body: "not an upload sample"},
	}
	if got := aggregateUploadTime(samples).(string); got != "2s" {
		t.Errorf("expected '2s' (wrong-type skipped), got %v", got)
	}
}
