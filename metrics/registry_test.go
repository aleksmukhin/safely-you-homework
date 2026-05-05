package metrics

import (
	"testing"

	"safely-you-homework/adapters"
)

func TestRegisterAndLookup(t *testing.T) {
	name := adapters.MetricName("registry_test_metric_xyz")
	def := MetricDef{Name: name, JSONKey: "test_key"}
	Register(def)
	defer delete(registry, name)

	got, ok := Lookup(name)
	if !ok {
		t.Fatal("expected to find registered metric")
	}
	if got.JSONKey != "test_key" {
		t.Errorf("expected JSONKey=test_key, got %s", got.JSONKey)
	}
}

func TestLookupUnknown(t *testing.T) {
	if _, ok := Lookup("nonexistent_metric_xyz"); ok {
		t.Error("expected ok=false for unknown metric")
	}
}

func TestAllSorted(t *testing.T) {
	all := All()
	for i := 1; i < len(all); i++ {
		if all[i].Name < all[i-1].Name {
			t.Errorf("All() should be sorted by name: %s < %s at index %d", all[i].Name, all[i-1].Name, i)
		}
	}
}

func TestAllIncludesInitRegistrations(t *testing.T) {
	expected := map[adapters.MetricName]bool{
		"heartbeat":   false,
		"upload_time": false,
		"firmware":    false,
	}
	for _, def := range All() {
		if _, want := expected[def.Name]; want {
			expected[def.Name] = true
		}
	}
	for name, found := range expected {
		if !found {
			t.Errorf("expected %q to be registered via init()", name)
		}
	}
}
