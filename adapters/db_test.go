package adapters

import (
	"sync"
	"testing"
)

func newTestDb(deviceIDs ...string) *DeviceDb {
	db := &DeviceDb{Devices: map[string]*DeviceData{}}
	for _, id := range deviceIDs {
		db.Devices[id] = &DeviceData{
			Metrics: map[MetricName][]StoredSample{},
		}
	}
	return db
}

func TestDeviceExists(t *testing.T) {
	db := newTestDb("known")
	if !db.DeviceExists("known") {
		t.Error("expected known device to exist")
	}
	if db.DeviceExists("unknown") {
		t.Error("expected unknown device to not exist")
	}
}

func TestAddSampleAndGetSamples(t *testing.T) {
	db := newTestDb("d1")
	if err := db.AddSample("d1", "heartbeat", "body1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := db.AddSample("d1", "heartbeat", "body2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	samples, ok := db.GetSamples("d1", "heartbeat")
	if !ok {
		t.Fatal("expected device to be found")
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 samples, got %d", len(samples))
	}
	if samples[0].Body != "body1" || samples[1].Body != "body2" {
		t.Errorf("samples in unexpected order: %+v", samples)
	}
	if samples[0].IngestedAt.IsZero() {
		t.Error("expected IngestedAt to be stamped by server")
	}
}

func TestAddSampleUnknownDevice(t *testing.T) {
	db := newTestDb()
	if err := db.AddSample("nope", "heartbeat", "body"); err == nil {
		t.Fatal("expected error for unknown device")
	}
}

func TestGetSamplesUnknownDevice(t *testing.T) {
	db := newTestDb()
	if _, ok := db.GetSamples("nope", "heartbeat"); ok {
		t.Error("expected ok=false for unknown device")
	}
}

func TestGetSamplesUnknownMetric(t *testing.T) {
	db := newTestDb("d1")
	samples, ok := db.GetSamples("d1", "nonexistent")
	if !ok {
		t.Fatal("expected ok=true for known device")
	}
	if len(samples) != 0 {
		t.Errorf("expected 0 samples for unknown metric, got %d", len(samples))
	}
}

func TestGetSamplesReturnsCopy(t *testing.T) {
	db := newTestDb("d1")
	if err := db.AddSample("d1", "heartbeat", "original"); err != nil {
		t.Fatal(err)
	}

	samples, _ := db.GetSamples("d1", "heartbeat")
	samples[0].Body = "mutated"

	fresh, _ := db.GetSamples("d1", "heartbeat")
	if fresh[0].Body != "original" {
		t.Error("GetSamples should return a copy; caller mutation leaked into storage")
	}
}

func TestConcurrentWritesAndReads(t *testing.T) {
	db := newTestDb("d1")
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = db.AddSample("d1", "heartbeat", "x")
		}()
		go func() {
			defer wg.Done()
			_, _ = db.GetSamples("d1", "heartbeat")
		}()
	}
	wg.Wait()
	samples, _ := db.GetSamples("d1", "heartbeat")
	if len(samples) != 100 {
		t.Errorf("expected 100 samples after concurrent writes, got %d", len(samples))
	}
}
