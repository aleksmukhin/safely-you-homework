package adapters

import (
	"fmt"
	"sync"
)

type DeviceDb struct {
	sync.RWMutex
	Devices map[string]*DeviceData
}

type DeviceData struct {
	Heartbeats []DeviceHeartbeat
	Stats      []DeviceStats
}

type DeviceStats struct {
	SentAt     string `json:"sent_at"`
	UploadTime int    `json:"upload_time"`
}

type DeviceHeartbeat struct {
	SentAt string `json:"sent_at"`
}

func (db *DeviceDb) AddDeviceHeartbeat(deviceID, sentAt string) error {
	db.Lock()
	defer db.Unlock()

	device, exists := db.Devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found")
	}

	device.Heartbeats = append(device.Heartbeats, DeviceHeartbeat{SentAt: sentAt})
	return nil
}

func (db *DeviceDb) AddDeviceStats(deviceID, sentAt string, uploadTime int) error {
	db.Lock()
	defer db.Unlock()

	device, exists := db.Devices[deviceID]
	if !exists {
		return fmt.Errorf("device not found")
	}

	device.Stats = append(device.Stats, DeviceStats{SentAt: sentAt, UploadTime: uploadTime})
	return nil
}

func (db *DeviceDb) GetDeviceData(deviceID string) ([]DeviceHeartbeat, []DeviceStats, bool) {
	db.RLock()
	defer db.RUnlock()

	device, exists := db.Devices[deviceID]
	if !exists {
		return nil, nil, false
	}

	heartbeats := make([]DeviceHeartbeat, len(device.Heartbeats))
	copy(heartbeats, device.Heartbeats)
	stats := make([]DeviceStats, len(device.Stats))
	copy(stats, device.Stats)
	return heartbeats, stats, true
}
