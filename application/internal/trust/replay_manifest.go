package trust

import (
	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/replay"
)

// buildReplayManifest fills replay.yaml content from a verification run (spec §21).
func buildReplayManifest(repoRoot string, scope VerificationScope, checks []VerificationCheck, cfg *config.Config) replay.Manifest {
	m := replay.Manifest{
		TrustID: scope.TrustID,
		Flow:    scope.Flow,
		Branch:  scope.Branch,
		Checks:  executedCheckTypes(checks),
	}
	if commit, err := bootstrap.GitHead(repoRoot); err == nil {
		m.RepoCommit = commit
	}
	if cfg != nil {
		for _, vc := range cfg.Validation.Commands {
			if vc.Command == "" {
				continue
			}
			m.Commands = append(m.Commands, vc.Command)
		}
	}
	return m
}

func executedCheckTypes(checks []VerificationCheck) []string {
	if len(checks) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(checks))
	out := make([]string, 0, len(checks))
	for _, c := range checks {
		t := string(c.Type)
		if t == "" {
			continue
		}
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
