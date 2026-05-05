package metrics

import (
	"github.com/gin-gonic/gin"
	"safely-you-homework/adapters"
)

type firmwareBody struct {
	Version string `json:"version" binding:"required"`
}

func init() {
	Register(MetricDef{
		Name:    "firmware",
		JSONKey: "firmware",
		Bind: func(c *gin.Context) (any, error) {
			var b firmwareBody
			if err := c.ShouldBindJSON(&b); err != nil {
				return nil, err
			}
			return b, nil
		},
		Aggregate: func(samples []adapters.StoredSample) any {
			for i := len(samples) - 1; i >= 0; i-- {
				if b, ok := samples[i].Body.(firmwareBody); ok {
					return b.Version
				}
			}
			return nil
		},
	})
}
