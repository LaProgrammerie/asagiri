package telemetry

import (
	"testing"
	"time"
)

func TestRunMetricFields(t *testing.T) {
	m := RunMetric{
		RunID:     "r1",
		Feature:   "f",
		TaskID:    "t",
		StartedAt: time.Now().UTC(),
		Status:    "done",
	}
	if m.RunID != "r1" || m.Status != "done" {
		t.Fatalf("%+v", m)
	}
}
