package replay

import (
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// ViewModel contains replay explorer data.
type ViewModel struct {
	Replay  bus.ReplayPackageResult
	ShowCLI bool
}

// Render returns replay timeline content.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Replay\n")
	b.WriteString("ID: " + value(vm.Replay.ReplayID, "-") + "\n")
	if !vm.Replay.CreatedAt.IsZero() {
		b.WriteString("Created: " + vm.Replay.CreatedAt.Format(time.RFC3339) + "\n")
	}
	if vm.Replay.RepoBranch != "" || vm.Replay.RepoCommit != "" {
		b.WriteString(fmt.Sprintf("Repo: %s @ %s\n", value(vm.Replay.RepoBranch, "-"), shortCommit(vm.Replay.RepoCommit)))
	}
	b.WriteString("Mode: " + value(vm.Replay.Mode, "full") + "\n")
	if vm.Replay.Warning != "" {
		b.WriteString("Warning: " + vm.Replay.Warning + "\n")
	}
	if len(vm.Replay.Artifacts) > 0 {
		b.WriteString("\nArtifacts\n")
		for _, artifact := range vm.Replay.Artifacts {
			b.WriteString("- " + artifact + "\n")
		}
	}
	b.WriteString("\nTimeline\n")
	if len(vm.Replay.Timeline) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, event := range vm.Replay.Timeline {
			label := event.Type
			if event.Artifact != "" {
				label += " artifact=" + event.Artifact
			}
			b.WriteString(fmt.Sprintf("- %s %s\n", event.Time.Format("15:04:05"), label))
		}
	}
	b.WriteString("\nActions\n")
	b.WriteString("- jump to event (stub)\n")
	b.WriteString("- inspect artifact (stub)\n")
	b.WriteString("- compare run: asa replay compare <replay-a> <replay-b>\n")
	b.WriteString("- replay offline: asa replay run " + value(vm.Replay.ReplayID, "<replay-id>") + " --offline\n")
	if vm.ShowCLI {
		b.WriteString("- open replay: asa replay open " + value(vm.Replay.ReplayID, "<replay-id>"))
	} else {
		b.WriteString("- explain divergence: asa replay explain <replay-a> <replay-b>")
	}
	return strings.TrimRight(b.String(), "\n")
}

func value(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func shortCommit(commit string) string {
	c := strings.TrimSpace(commit)
	if len(c) > 12 {
		return c[:12]
	}
	return c
}
