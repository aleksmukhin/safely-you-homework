package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"safely-you-homework/adapters"
)

type heartbeatBody struct {
	SentAt time.Time `json:"sent_at" binding:"required"`
}

func init() {
	Register(MetricDef{
		Name:    "heartbeat",
		JSONKey: "uptime",
		Bind: func(c *gin.Context) (any, error) {
			var b heartbeatBody
			if err := c.ShouldBindJSON(&b); err != nil {
				return nil, err
			}
			return b, nil
		},
		Aggregate: aggregateHeartbeats,
	})
}

func aggregateHeartbeats(samples []adapters.StoredSample) any {
	if len(samples) == 0 {
		return float64(0)
	}

	minutes := make(map[int64]struct{}, len(samples))
	var first, last time.Time
	for _, s := range samples {
		b, ok := s.Body.(heartbeatBody)
		if !ok {
			continue
		}
		bucket := b.SentAt.Truncate(time.Minute).Unix() / 60
		minutes[bucket] = struct{}{}
		if first.IsZero() || b.SentAt.Before(first) {
			first = b.SentAt
		}
		if b.SentAt.After(last) {
			last = b.SentAt
		}
	}

	if len(minutes) == 0 {
		return float64(0)
	}

	totalMinutes := int64(last.Sub(first) / time.Minute)
	if totalMinutes == 0 {
		return float64(100)
	}
	return (float64(len(minutes)) / float64(totalMinutes)) * 100
}
