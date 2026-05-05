package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"safely-you-homework/adapters"
	_ "safely-you-homework/metrics"
)

func setupTestServer(deviceIDs ...string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	db := &adapters.DeviceDb{Devices: map[string]*adapters.DeviceData{}}
	for _, id := range deviceIDs {
		db.Devices[id] = &adapters.DeviceData{
			Metrics: map[adapters.MetricName][]adapters.StoredSample{},
		}
	}
	handler := &DeviceHandler{Db: db}

	r := gin.New()
	v2 := r.Group("/api/v2")
	{
		v2.POST("/devices/:device_id/metrics/:metric_name", handler.PostMetric)
		v2.GET("/devices/:device_id/stats", handler.GetDeviceStats)
	}
	return r
}

func do(t *testing.T, r *gin.Engine, method, path, body string) (int, string) {
	t.Helper()
	var req *http.Request
	if body != "" {
		req = httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code, w.Body.String()
}

func TestPostHeartbeatHappyPath(t *testing.T) {
	r := setupTestServer("dev1")
	code, _ := do(t, r, "POST", "/api/v2/devices/dev1/metrics/heartbeat", `{"sent_at":"2026-05-04T10:00:00Z"}`)
	if code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestPostUploadTimeHappyPath(t *testing.T) {
	r := setupTestServer("dev1")
	code, _ := do(t, r, "POST", "/api/v2/devices/dev1/metrics/upload_time",
		`{"sent_at":"2026-05-04T10:00:00Z","upload_time":2000000000}`)
	if code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", code)
	}
}

func TestPostFirmwareWithoutSentAt(t *testing.T) {
	r := setupTestServer("dev1")
	code, _ := do(t, r, "POST", "/api/v2/devices/dev1/metrics/firmware", `{"version":"2.0.0"}`)
	if code != http.StatusNoContent {
		t.Errorf("expected 204 (firmware metric does not require sent_at), got %d", code)
	}
}

func TestPostUnknownMetric(t *testing.T) {
	r := setupTestServer("dev1")
	code, body := do(t, r, "POST", "/api/v2/devices/dev1/metrics/bogus", `{}`)
	if code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
	if !strings.Contains(body, `"msg"`) || !strings.Contains(body, "Unknown metric") {
		t.Errorf("expected msg about unknown metric, got %s", body)
	}
}

func TestPostUnknownDevice(t *testing.T) {
	r := setupTestServer()
	code, body := do(t, r, "POST", "/api/v2/devices/nope/metrics/heartbeat",
		`{"sent_at":"2026-05-04T10:00:00Z"}`)
	if code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
	if !strings.Contains(body, "Device not found") {
		t.Errorf("expected device-not-found msg, got %s", body)
	}
}

func TestPostMissingRequiredField(t *testing.T) {
	r := setupTestServer("dev1")
	code, _ := do(t, r, "POST", "/api/v2/devices/dev1/metrics/heartbeat", `{}`)
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing sent_at, got %d", code)
	}
}

func TestPostInvalidJSON(t *testing.T) {
	r := setupTestServer("dev1")
	code, _ := do(t, r, "POST", "/api/v2/devices/dev1/metrics/heartbeat", `not json`)
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", code)
	}
}

func TestGetCompositeStats(t *testing.T) {
	r := setupTestServer("dev1")
	do(t, r, "POST", "/api/v2/devices/dev1/metrics/heartbeat", `{"sent_at":"2026-05-04T10:00:00Z"}`)
	do(t, r, "POST", "/api/v2/devices/dev1/metrics/upload_time",
		`{"sent_at":"2026-05-04T10:00:00Z","upload_time":2000000000}`)
	do(t, r, "POST", "/api/v2/devices/dev1/metrics/firmware", `{"version":"3.0.0"}`)

	code, body := do(t, r, "GET", "/api/v2/devices/dev1/stats", "")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", code, body)
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if _, ok := resp["uptime"]; !ok {
		t.Error("expected uptime key in composite response")
	}
	if _, ok := resp["avg_upload_time"]; !ok {
		t.Error("expected avg_upload_time key in composite response")
	}
	if resp["firmware"] != "3.0.0" {
		t.Errorf("expected firmware=3.0.0, got %v", resp["firmware"])
	}
}

func TestGetFilteredSingleMetric(t *testing.T) {
	r := setupTestServer("dev1")
	do(t, r, "POST", "/api/v2/devices/dev1/metrics/firmware", `{"version":"4.5.6"}`)

	code, body := do(t, r, "GET", "/api/v2/devices/dev1/stats?metric=firmware", "")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp) != 1 {
		t.Errorf("expected 1 key in filtered response, got %v", resp)
	}
	if resp["firmware"] != "4.5.6" {
		t.Errorf("expected firmware=4.5.6, got %v", resp["firmware"])
	}
}

func TestGetFilteredMultipleMetrics(t *testing.T) {
	r := setupTestServer("dev1")
	do(t, r, "POST", "/api/v2/devices/dev1/metrics/heartbeat", `{"sent_at":"2026-05-04T10:00:00Z"}`)
	do(t, r, "POST", "/api/v2/devices/dev1/metrics/firmware", `{"version":"7.7.7"}`)

	code, body := do(t, r, "GET", "/api/v2/devices/dev1/stats?metric=heartbeat,firmware", "")
	if code != http.StatusOK {
		t.Fatalf("expected 200, got %d", code)
	}

	var resp map[string]any
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(resp) != 2 {
		t.Errorf("expected 2 keys in filtered response, got %v", resp)
	}
}

func TestGetUnknownDevice(t *testing.T) {
	r := setupTestServer()
	code, _ := do(t, r, "GET", "/api/v2/devices/nope/stats", "")
	if code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
}

func TestGetUnknownMetricInQuery(t *testing.T) {
	r := setupTestServer("dev1")
	code, body := do(t, r, "GET", "/api/v2/devices/dev1/stats?metric=bogus", "")
	if code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", code)
	}
	if !strings.Contains(body, "Unknown metric") {
		t.Errorf("expected msg about unknown metric, got %s", body)
	}
}
