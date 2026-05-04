package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"safely-you-homework/adapters"
)

type DeviceHandler struct {
	Db *adapters.DeviceDb
}

type DeviceStatsResponse struct {
	Uptime        float64 `json:"uptime"`
	AvgUploadTime string  `json:"avg_upload_time"`
}

func (h *DeviceHandler) GetDeviceStats(c *gin.Context) {
	
	deviceID := c.Param("device_id")
	
	heartbeats, stats, ok := h.Db.GetDeviceData(deviceID)
	
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"msg": "Device not found"})
		return
	}

	c.JSON(http.StatusOK, DeviceStatsResponse{
		Uptime:        computeUptime(heartbeats),
		AvgUploadTime: computeAvgUploadTime(stats),
	})
}

func (h *DeviceHandler) PostDeviceStats(c *gin.Context) {
	deviceID := c.Param("device_id")
	
	var stats adapters.DeviceStats
	
	if err := c.ShouldBindJSON(&stats); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid JSON"})
		return
	}
	
	if err := h.Db.AddDeviceStats(deviceID, stats.SentAt, stats.UploadTime); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "Device not found"})
		return
	}
	
	c.Status(http.StatusNoContent)
}

func (h *DeviceHandler) PostDeviceHeartbeat(c *gin.Context) {
	deviceID := c.Param("device_id")

	var hb adapters.DeviceHeartbeat
	
	if err := c.ShouldBindJSON(&hb); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "Invalid JSON"})
		return
	}
	
	if err := h.Db.AddDeviceHeartbeat(deviceID, hb.SentAt); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "Device not found"})
		return
	}
	
	c.Status(http.StatusNoContent)
}

func computeUptime(heartbeats []adapters.DeviceHeartbeat) float64 {
	if len(heartbeats) == 0 {
		return 0
	}

	minutes := make(map[int64]struct{}, len(heartbeats))
	
	var first, last time.Time
	
	for _, hb := range heartbeats {
		t, err := time.Parse(time.RFC3339, hb.SentAt)
		if err != nil {
			continue
		}
		bucket := t.Truncate(time.Minute).Unix() / 60
		minutes[bucket] = struct{}{}
		if first.IsZero() || t.Before(first) {
			first = t
		}
		if t.After(last) {
			last = t
		}
	}

	if len(minutes) == 0 {
		return 0
	}

	totalMinutes := int64(last.Sub(first) / time.Minute)

	if totalMinutes == 0 {
		return 100
	}

	return (float64(len(minutes)) / float64(totalMinutes)) * 100
}

func computeAvgUploadTime(stats []adapters.DeviceStats) string {
	if len(stats) == 0 {
		return time.Duration(0).String()
	}
	
	var total int64
	
	for _, s := range stats {
		total += int64(s.UploadTime)
	}
	
	avg := total / int64(len(stats))
	
	return time.Duration(avg).String()
}
