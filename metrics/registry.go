package metrics

import (
	"sort"

	"github.com/gin-gonic/gin"
	"safely-you-homework/adapters"
)

type MetricDef struct {
	Name      adapters.MetricName
	JSONKey   string
	Bind      func(*gin.Context) (any, error)
	Aggregate func([]adapters.StoredSample) any
}

var registry = map[adapters.MetricName]MetricDef{}

func Register(def MetricDef) {
	registry[def.Name] = def
}

func Lookup(name adapters.MetricName) (MetricDef, bool) {
	def, ok := registry[name]
	return def, ok
}

func All() []MetricDef {
	out := make([]MetricDef, 0, len(registry))
	for _, def := range registry {
		out = append(out, def)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out
}
