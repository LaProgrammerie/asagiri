package worktree

import (
	"context"
	"fmt"
	"os"
	osExec "os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var nonAlnum = regexp.MustCompile(`[^a-z0-9._-]+`)

// Manager handles git worktree lifecycle for task execution.
type Manager struct {
	RepoRoot      string
	BasePath      string
	BranchPrefix  string
	DefaultBranch string
	DryRun        bool
}

func New(repoRoot, basePath, branchPrefix, defaultBranch string, dryRun bool) *Manager {
	return &Manager{
		RepoRoot:      repoRoot,
		BasePath:      basePath,
		BranchPrefix:  branchPrefix,
		DefaultBranch: defaultBranch,
		DryRun:        dryRun,
	}
}

func sanitize(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	s = strings.ReplaceAll(s, " ", "-")
	s = nonAlnum.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "task"
	}
	return s
}

// Create creates one worktree and dedicated branch for a task.
func (m *Manager) Create(ctx context.Context, feature, taskID string) (string, string, error) {
	featureSlug := sanitize(feature)
	taskSlug := sanitize(taskID)
	branch := fmt.Sprintf("%s/%s-%s", sanitize(m.BranchPrefix), featureSlug, taskSlug)
	path := filepath.Join(m.BasePath, featureSlug, taskSlug)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", "", fmt.Errorf("create worktree parent: %w", err)
	}
	if m.DryRun {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", "", fmt.Errorf("dry-run mkdir worktree: %w", err)
		}
		return path, branch, nil
	}

	args := []string{
		"-C", m.RepoRoot,
		"worktree", "add", "-b", branch, path,
	}
	if m.DefaultBranch != "" {
		args = append(args, m.DefaultBranch)
	}
	cmd := osExec.CommandContext(ctx, "git", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", "", fmt.Errorf("git worktree add: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return path, branch, nil
}

// Remove removes a worktree path.
func (m *Manager) Remove(ctx context.Context, path string) error {
	if path == "" {
		return fmt.Errorf("worktree path vide")
	}
	if m.DryRun {
		return os.RemoveAll(path)
	}

	cmd := osExec.CommandContext(ctx, "git", "-C", m.RepoRoot, "worktree", "remove", "--force", path)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git worktree remove: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
