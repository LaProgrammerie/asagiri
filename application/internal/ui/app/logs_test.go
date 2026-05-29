package app

import (
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestLogLinesFromEventsLevels(t *testing.T) {
	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	lines := logLinesFromEvents([]bus.EventSummary{
		{Type: "graph.generated", CreatedAt: now},
		{Type: "trust.failed", CreatedAt: now},
	})
	require.Len(t, lines, 2)
	require.Equal(t, "info", lines[0].Level)
	require.Equal(t, "error", lines[1].Level)
}
