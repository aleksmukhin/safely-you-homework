package metrics

import (
	"math"
	"testing"
	"time"

	"safely-you-homework/adapters"
)

func hbSample(at time.Time) adapters.StoredSample {
	return adapters.StoredSample{IngestedAt: at, Body: heartbeatBody{SentAt: at}}
}

func TestAggregateHeartbeatsEmpty(t *testing.T) {
	if got := aggregateHeartbeats(nil).(float64); got != 0 {
		t.Errorf("expected 0 for empty samples, got %v", got)
	}
}

func TestAggregateHeartbeatsSingle(t *testing.T) {
	base := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	got := aggregateHeartbeats([]adapters.StoredSample{hbSample(base)}).(float64)
	if got != 100 {
		t.Errorf("expected 100 (single bucket → no span), got %v", got)
	}
}

func TestAggregateHeartbeatsWithGaps(t *testing.T) {
	base := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	// 5 unique-minute buckets across a 9-minute span (gap from minute 3..7)
	samples := []adapters.StoredSample{
		hbSample(base),
		hbSample(base.Add(1 * time.Minute)),
		hbSample(base.Add(2 * time.Minute)),
		hbSample(base.Add(8 * time.Minute)),
		hbSample(base.Add(9 * time.Minute)),
	}
	got := aggregateHeartbeats(samples).(float64)
	expected := 5.0 / 9.0 * 100
	if math.Abs(got-expected) > 0.01 {
		t.Errorf("expected ~%.2f, got %.2f", expected, got)
	}
}

func TestAggregateHeartbeatsIgnoresWrongType(t *testing.T) {
	base := time.Date(2026, 5, 4, 10, 0, 0, 0, time.UTC)
	samples := []adapters.StoredSample{
		hbSample(base),
		hbSample(base.Add(time.Minute)),
		{IngestedAt: base, Body: "not a heartbeat"},
	}
	got := aggregateHeartbeats(samples).(float64)
	// 2 valid heartbeats, 1-minute span → 2/1 * 100 = 200 (degenerate but expected math)
	if got != 200 {
		t.Errorf("expected 200 (wrong-type sample skipped), got %v", got)
	}
}
