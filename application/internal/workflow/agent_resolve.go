package workflow

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/agentresolve"
)

func (s *Service) resolveAgentPrompt(phase agentresolve.Phase, agentKey, feature, taskID, runID, legacyPrompt string, contextFiles []string) (string, error) {
	resolved, err := agentresolve.Resolve(agentresolve.Params{
		RepoRoot:     s.repoRoot,
		Config:       s.cfg,
		Phase:        phase,
		AgentKey:     agentKey,
		RunID:        runID,
		Feature:      feature,
		TaskID:       taskID,
		LegacyPrompt: legacyPrompt,
		ContextFiles: contextFiles,
	})
	if err != nil {
		return "", err
	}
	return resolved.Prompt, nil
}

func (s *Service) resolveGatePrompt(phase agentresolve.Phase, agentKey, feature, taskID, runID, legacyPrompt string, contextFiles []string) (string, error) {
	prompt, err := s.resolveAgentPrompt(phase, agentKey, feature, taskID, runID, legacyPrompt, contextFiles)
	if err != nil {
		return "", fmt.Errorf("%s prompt resolve: %w", phase, err)
	}
	return prompt, nil
}
