package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"safely-you-homework/adapters"
	"safely-you-homework/metrics"
)

type DeviceHandler struct {
	Db *adapters.DeviceDb
}

func (h *DeviceHandler) PostMetric(c *gin.Context) {
	deviceID := c.Param("device_id")
	name := adapters.MetricName(c.Param("metric_name"))

	def, ok := metrics.Lookup(name)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"msg": "Unknown metric"})
		return
	}

	body, err := def.Bind(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": err.Error()})
		return
	}

	if err := h.Db.AddSample(deviceID, name, body); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "Device not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *DeviceHandler) GetDeviceStats(c *gin.Context) {
	deviceID := c.Param("device_id")

	if !h.Db.DeviceExists(deviceID) {
		c.JSON(http.StatusNotFound, gin.H{"msg": "Device not found"})
		return
	}

	var defs []metrics.MetricDef
	if q := c.Query("metric"); q != "" {
		for _, raw := range strings.Split(q, ",") {
			name := strings.TrimSpace(raw)
			def, ok := metrics.Lookup(adapters.MetricName(name))
			if !ok {
				c.JSON(http.StatusNotFound, gin.H{"msg": fmt.Sprintf("Unknown metric: %s", name)})
				return
			}
			defs = append(defs, def)
		}
	} else {
		defs = metrics.All()
	}

	response := map[string]any{}
	for _, def := range defs {
		samples, _ := h.Db.GetSamples(deviceID, def.Name)
		response[def.JSONKey] = def.Aggregate(samples)
	}

	c.JSON(http.StatusOK, response)
}
