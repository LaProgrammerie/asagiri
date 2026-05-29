package coordination

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

const defaultAgentRunsDir = ".asagiri/agent-runs"

// NodeRunRequest is input for a single coordinated node execution.
type NodeRunRequest struct {
	RepoRoot     string
	Graph        executiongraph.ExecutionGraph
	Node         executiongraph.GraphNode
	Assignment   AgentAssignment
	WorkingDir   string
	WorktreePath string
}

// NodeAgentRunner executes one graph node (real agent subprocess or test stub).
type NodeAgentRunner interface {
	Run(ctx context.Context, req NodeRunRequest) error
}

// MarkerNodeAgentRunner writes a completion marker under .asagiri/agent-runs/ (testable stub).
type MarkerNodeAgentRunner struct {
	RunsDir string
}

// Run creates <runsDir>/<graphID>/<nodeID>.done in the working directory tree.
func (m *MarkerNodeAgentRunner) Run(_ context.Context, req NodeRunRequest) error {
	dir := m.RunsDir
	if dir == "" {
		dir = defaultAgentRunsDir
	}
	base := req.RepoRoot
	if strings.TrimSpace(req.WorkingDir) != "" {
		base = req.WorkingDir
	}
	if base == "" {
		return fmt.Errorf("%w: repo root required for marker agent", ErrInvalidAssignment)
	}
	markerDir := filepath.Join(base, dir, req.Graph.ID)
	if err := os.MkdirAll(markerDir, 0o755); err != nil {
		return fmt.Errorf("marker agent mkdir: %w", err)
	}
	marker := filepath.Join(markerDir, req.Node.ID+".done")
	body := fmt.Sprintf("agent=%s role=%s isolation=%s\n", req.Assignment.AgentRef, req.Assignment.Role, req.Assignment.Isolation)
	if err := os.WriteFile(marker, []byte(body), 0o644); err != nil {
		return fmt.Errorf("marker agent write: %w", err)
	}
	return nil
}

// NodeExecutor returns an executiongraph.NodeExecutor with worktree isolation and agent stub.
func NodeExecutor(repoRoot string, cfg *config.Config) executiongraph.NodeExecutor {
	runner := &MarkerNodeAgentRunner{}
	return func(ctx context.Context, graph executiongraph.ExecutionGraph, node executiongraph.GraphNode, asg executiongraph.CoordinationAssignment) error {
		assignment := AgentAssignment{
			NodeID:    asg.NodeID,
			AgentRef:  asg.AgentRef,
			Role:      AgentRole(asg.Role),
			Isolation: IsolationMode(asg.Isolation),
			ProfileID: asg.ProfileID,
		}
		workDir := repoRoot
		var cleanup func()
		if assignment.Isolation == IsolationIsolatedWorktree {
			branch := WorktreeBranch(cfg, graph.ID, node.ID)
			path, release, err := EnsureWorktree(ctx, repoRoot, graph.ID, node.ID, branch)
			if err != nil {
				return err
			}
			cleanup = release
			workDir = path
		}
		if cleanup != nil {
			defer cleanup()
		}
		return runner.Run(ctx, NodeRunRequest{
			RepoRoot:     repoRoot,
			Graph:        graph,
			Node:         node,
			Assignment:   assignment,
			WorkingDir:   workDir,
			WorktreePath: workDir,
		})
	}
}

// RunOptions wires coordinator and node executor for graph runs (spec-my-D D-FULL).
func RunOptions(cfg *config.Config, repoRoot string, emitter *CoordinationEmitter) executiongraph.RunOptions {
	coord := NewDefaultCoordinator(cfg, repoRoot, emitter)
	return executiongraph.RunOptions{
		Coordinator: GraphCoordinator(coord),
		NodeExecutor: NodeExecutor(repoRoot, cfg),
	}
}

// WorktreeBranch builds a dedicated branch name for a coordinated node worktree.
func WorktreeBranch(cfg *config.Config, graphID, nodeID string) string {
	prefix := config.DefaultBranchPrefix
	if cfg != nil && cfg.Worktrees.BranchPrefix != "" {
		prefix = cfg.Worktrees.BranchPrefix
	}
	return fmt.Sprintf("%s/coord/%s/%s", sanitizePathSegment(prefix), sanitizePathSegment(graphID), sanitizePathSegment(nodeID))
}
