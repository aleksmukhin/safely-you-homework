package main

import (
	"encoding/csv"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"safely-you-homework/handlers"
	"safely-you-homework/adapters"
)

func main() {
	r := gin.Default()

	// load devices into a permament in-memory storage
	db := loadDbFromCSV("devices.csv")

	// create handlers
	deviceHandler := &handlers.DeviceHandler{
		Db: db,
	}

	// register routes
	v2 := r.Group("/api/v2")
	{
		v2.POST("/devices/:device_id/metrics/:metric_name", deviceHandler.PostMetric)
		v2.GET("/devices/:device_id/stats", deviceHandler.GetDeviceStats)
	}

	r.Run(":6733")
}

func loadDbFromCSV(filename string) *adapters.DeviceDb {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	
	// Skip header
	if _, err := reader.Read(); err != nil {
		log.Fatal(err)
	}

	db := &adapters.DeviceDb{
		Devices: make(map[string]*adapters.DeviceData),
	}

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range records {
		deviceID := row[0] 
		db.Devices[deviceID] = &adapters.DeviceData{
			Metrics: map[adapters.MetricName][]adapters.StoredSample{},
		}
	}

	return db
}