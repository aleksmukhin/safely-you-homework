package adapters

import (
	"fmt"
	"sync"
	"time"
)

type MetricName string

type StoredSample struct {
	IngestedAt time.Time
	Body       any
}

type DeviceDb struct {
	sync.RWMutex
	Devices map[string]*DeviceData
}

type DeviceData struct {
	Metrics map[MetricName][]StoredSample
}

func (db *DeviceDb) DeviceExists(deviceID string) bool {
	db.RLock()
	defer db.RUnlock()
	_, ok := db.Devices[deviceID]
	return ok
}

func (db *DeviceDb) AddSample(deviceID string, name MetricName, body any) error {
	db.Lock()
	defer db.Unlock()

	device, exists := db.Devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found")
	}
	if device.Metrics == nil {
		device.Metrics = map[MetricName][]StoredSample{}
	}
	device.Metrics[name] = append(device.Metrics[name], StoredSample{
		IngestedAt: time.Now(),
		Body:       body,
	})
	return nil
}

func (db *DeviceDb) GetSamples(deviceID string, name MetricName) ([]StoredSample, bool) {
	db.RLock()
	defer db.RUnlock()

	device, exists := db.Devices[deviceID]
	if !exists {
		return nil, false
	}
	samples := device.Metrics[name]
	out := make([]StoredSample, len(samples))
	copy(out, samples)
	return out, true
}
