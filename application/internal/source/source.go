package source

import (
	"context"
	"time"
)

// Source abstracts external spec providers (specv2 §7.1).
type Source interface {
	Name() string
	List(ctx context.Context) ([]SourceItem, error)
	Fetch(ctx context.Context, ref SourceRef) (SourceDocument, error)
	Sync(ctx context.Context, ref SourceRef, dest LocalSpecPath, opts SyncOptions) (SyncResult, error)
}

// SourceRef identifies a remote document.
type SourceRef struct {
	ID   string
	URL  string
	Name string
}

// SourceItem is an inbox entry.
type SourceItem struct {
	Ref       SourceRef
	Feature   string
	Status    string
	UpdatedAt time.Time
	PathHint  string
}

// SourceDocument is fetched remote content.
type SourceDocument struct {
	Feature         string
	Title           string
	Markdown        string
	TasksYAML       string
	Status          string
	RemoteUpdatedAt time.Time
	Ref             SourceRef
}

// LocalSpecPath is the destination under the repo.
type LocalSpecPath struct {
	Root    string
	Feature string
}

// SyncOptions controls overwrite behaviour.
type SyncOptions struct {
	Force       bool
	Interactive bool
}

// SyncResult reports sync outcome.
type SyncResult struct {
	Feature      string
	Path         string
	Overwritten  bool
	Conflict     bool
	NeedsConfirm bool
}
