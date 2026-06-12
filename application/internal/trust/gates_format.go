package trust

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// FormatGatesConfig renders configured verification gates for the terminal.
func FormatGatesConfig(v config.VerificationConfig) string {
	eval := NewGateEvaluator(&v)
	names := eval.ProfileNames()
	if len(names) == 0 {
		return "Verification gates: not configured\n"
	}

	var sb strings.Builder
	sb.WriteString("Verification gates\n")
	sb.WriteString("══════════════════\n")
	for _, name := range names {
		p := v.Gates[name]
		fmt.Fprintf(&sb, "\nProfile: %s\n", name)
		if len(p.MinConfidence) > 0 {
			sb.WriteString("  Min confidence:\n")
			keys := make([]string, 0, len(p.MinConfidence))
			for k := range p.MinConfidence {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Fprintf(&sb, "    %s: %.0f%%\n", k, p.MinConfidence[k]*100)
			}
		}
		if len(p.RequiredChecks) > 0 {
			sb.WriteString("  Required checks:\n")
			for _, c := range p.RequiredChecks {
				fmt.Fprintf(&sb, "    - %s\n", c)
			}
		}
	}
	if v.DefaultProfile != "" {
		fmt.Fprintf(&sb, "\nDefault profile: %s\n", v.DefaultProfile)
	}
	return sb.String()
}
