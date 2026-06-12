package workflow

import "context"

// SetDryRunForTest toggles dry-run mode (tests only).
func (s *Service) SetDryRunForTest(v bool) {
	s.dryRun = v
}

// SetEnrichGateAgentHookForTest overrides enrich gate agent output (tests only).
func (s *Service) SetEnrichGateAgentHookForTest(h func(context.Context, string) (string, error)) {
	s.enrichGateAgentHook = h
}

// SetVerifyEvidenceGateAgentHookForTest overrides verify_evidence gate agent output (tests only).
func (s *Service) SetVerifyEvidenceGateAgentHookForTest(h func(context.Context, string) (string, error)) {
	s.verifyEvidenceGateAgentHook = h
}
