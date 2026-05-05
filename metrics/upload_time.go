package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"safely-you-homework/adapters"
)

type uploadTimeBody struct {
	SentAt     time.Time `json:"sent_at" binding:"required"`
	UploadTime int       `json:"upload_time" binding:"required"`
}

func init() {
	Register(MetricDef{
		Name:    "upload_time",
		JSONKey: "avg_upload_time",
		Bind: func(c *gin.Context) (any, error) {
			var b uploadTimeBody
			if err := c.ShouldBindJSON(&b); err != nil {
				return nil, err
			}
			return b, nil
		},
		Aggregate: aggregateUploadTime,
	})
}

func aggregateUploadTime(samples []adapters.StoredSample) any {
	if len(samples) == 0 {
		return time.Duration(0).String()
	}
	var total int64
	var n int64
	for _, s := range samples {
		b, ok := s.Body.(uploadTimeBody)
		if !ok {
			continue
		}
		total += int64(b.UploadTime)
		n++
	}
	if n == 0 {
		return time.Duration(0).String()
	}
	return time.Duration(total / n).String()
}
