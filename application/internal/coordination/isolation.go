package coordination

import (
	"context"
	"fmt"
	"os"
	osExec "os/exec"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

// GitRunner runs git commands in a repository (injectable in tests).
type GitRunner func(ctx context.Context, repoRoot string, args ...string) (combinedOutput []byte, err error)

var defaultGitRunner GitRunner = func(ctx context.Context, repoRoot string, args ...string) ([]byte, error) {
	cmd := osExec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoRoot
	return cmd.CombinedOutput()
}

// gitRunner returns the active git runner (tests may replace defaultGitRunner).
func gitRunner() GitRunner {
	if defaultGitRunner == nil {
		return func(ctx context.Context, repoRoot string, args ...string) ([]byte, error) {
			cmd := osExec.CommandContext(ctx, "git", args...)
			cmd.Dir = repoRoot
			return cmd.CombinedOutput()
		}
	}
	return defaultGitRunner
}

// WorktreePath returns the isolated worktree directory for a graph node (spec-my-D §5).
func WorktreePath(repoRoot, graphID, nodeID string) (string, error) {
	if strings.TrimSpace(repoRoot) == "" {
		return "", fmt.Errorf("%w: repo root required", ErrInvalidAssignment)
	}
	if err := executiongraph.ValidateGraphID(graphID); err != nil {
		return "", err
	}
	safeNode := sanitizePathSegment(nodeID)
	if safeNode == "" {
		return "", fmt.Errorf("%w: node id required", ErrInvalidAssignment)
	}

	rel := filepath.Join(".asagiri", "worktrees", graphID, safeNode)
	clean := filepath.Clean(rel)
	if strings.HasPrefix(clean, "..") {
		return "", fmt.Errorf("%w: worktree path must not escape repo", ErrInvalidAssignment)
	}
	abs := filepath.Join(repoRoot, clean)
	relToRepo, err := filepath.Rel(repoRoot, abs)
	if err != nil || strings.HasPrefix(relToRepo, "..") {
		return "", fmt.Errorf("%w: worktree path must not escape repo", ErrInvalidAssignment)
	}
	return abs, nil
}

// EnsureWorktree creates or reuses a git worktree for isolated node execution (spec-my-D §5).
// cleanup removes the worktree when the caller finishes; it is safe to call multiple times.
func EnsureWorktree(ctx context.Context, repoRoot, graphID, nodeID, branch string) (string, func(), error) {
	wtPath, err := WorktreePath(repoRoot, graphID, nodeID)
	if err != nil {
		return "", nil, err
	}
	branch = strings.TrimSpace(branch)
	if branch == "" {
		return "", nil, fmt.Errorf("%w: branch required", ErrInvalidAssignment)
	}

	run := gitRunner()
	if exists, err := worktreeExists(ctx, run, repoRoot, wtPath); err != nil {
		return "", nil, err
	} else if exists {
		return wtPath, func() { _ = removeWorktree(ctx, run, repoRoot, wtPath) }, nil
	}

	if err := os.MkdirAll(filepath.Dir(wtPath), 0o755); err != nil {
		return "", nil, fmt.Errorf("create worktree parent: %w", err)
	}

	args := []string{"worktree", "add", "-b", branch, wtPath}
	if base, err := defaultBranchRef(ctx, run, repoRoot); err == nil && base != "" {
		args = append(args, base)
	}
	if out, err := run(ctx, repoRoot, args...); err != nil {
		if registered, listErr := worktreeExists(ctx, run, repoRoot, wtPath); listErr == nil && registered {
			return wtPath, func() { _ = removeWorktree(ctx, run, repoRoot, wtPath) }, nil
		}
		return "", nil, fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
	}

	return wtPath, func() { _ = removeWorktree(ctx, run, repoRoot, wtPath) }, nil
}

func worktreeExists(ctx context.Context, run GitRunner, repoRoot, path string) (bool, error) {
	out, err := run(ctx, repoRoot, "worktree", "list", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("git worktree list: %w: %s", err, strings.TrimSpace(string(out)))
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "worktree ") {
			wt := strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
			wtAbs, err := filepath.Abs(wt)
			if err != nil {
				continue
			}
			if wtAbs == abs {
				return true, nil
			}
		}
	}
	return false, nil
}

func removeWorktree(ctx context.Context, run GitRunner, repoRoot, path string) error {
	out, err := run(ctx, repoRoot, "worktree", "remove", "--force", path)
	if err != nil {
		return fmt.Errorf("git worktree remove: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func defaultBranchRef(ctx context.Context, run GitRunner, repoRoot string) (string, error) {
	out, err := run(ctx, repoRoot, "symbolic-ref", "--short", "HEAD")
	if err != nil {
		return "", err
	}
	ref := strings.TrimSpace(string(out))
	if ref == "" {
		return "", fmt.Errorf("empty HEAD")
	}
	return ref, nil
}

func sanitizePathSegment(id string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, strings.TrimSpace(id))
}
