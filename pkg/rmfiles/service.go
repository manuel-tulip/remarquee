// Package rmfiles provides a reusable, HTTP-agnostic service for uploading,
// managing, and searching reMarkable cloud files. It owns a single long-lived
// rmapi ApiCtx and serializes all mutations behind a mutex, because ApiCtx and
// its in-memory file tree are stateful and not safe for concurrent mutation.
package rmfiles

import (
	"context"
	"os/exec"
	"sync"
	"time"

	"github.com/go-go-golems/remarquee/pkg/rmcloud"
	"github.com/juruen/rmapi/api"
	"github.com/juruen/rmapi/model"
	"github.com/pkg/errors"
)

// Config configures a Service.
type Config struct {
	// Auth is passed to rmcloud when creating the ApiCtx (and on refresh).
	Auth rmcloud.AuthSettings
	// DefaultRemoteDir is the default upload destination (e.g. "/ai").
	DefaultRemoteDir string
	// IncludeTemplates exposes rmapi TemplateType entries (hidden by default).
	IncludeTemplates bool
	// Workers bounds concurrent conversions during an upload job.
	Workers int
}

// Status is the non-secret capability/state summary exposed to the UI.
type Status struct {
	Authenticated    bool      `json:"authenticated"`
	DocumentCount    int       `json:"documentCount"`
	LastRefresh      time.Time `json:"lastRefresh,omitempty"`
	DefaultRemoteDir string    `json:"defaultRemoteDir"`
	PandocAvailable  bool      `json:"pandocAvailable"`
}

// Service owns the cloud API context and derived state.
type Service struct {
	cfg Config

	mu          sync.Mutex
	apiCtx      api.ApiCtx
	index       []Entry
	lastRefresh time.Time
	jobs        *jobStore
}

// NewService wraps an already-created ApiCtx. Useful for tests and for callers
// that manage their own authentication.
func NewService(apiCtx api.ApiCtx, cfg Config) *Service {
	if cfg.Workers <= 0 {
		cfg.Workers = 2
	}
	if cfg.DefaultRemoteDir == "" {
		cfg.DefaultRemoteDir = "/"
	}
	s := &Service{cfg: cfg, apiCtx: apiCtx, jobs: newJobStore()}
	s.mu.Lock()
	s.reindexLocked()
	s.mu.Unlock()
	return s
}

// NewServiceFromCloud authenticates with the reMarkable cloud and builds a
// service around a fresh ApiCtx.
func NewServiceFromCloud(ctx context.Context, cfg Config) (*Service, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	auth := cfg.Auth
	auth.NonInteractive = true
	_, apiCtx, err := rmcloud.CreateApiCtx(ctx, auth)
	if err != nil {
		return nil, errors.Wrap(err, "create cloud api context")
	}
	return NewService(apiCtx, cfg), nil
}

// reindexLocked rebuilds the flat search index from the current tree. Callers
// must hold s.mu.
func (s *Service) reindexLocked() {
	if s.apiCtx == nil {
		s.index = nil
		return
	}
	s.index = WalkEntries(s.apiCtx.Filetree().Root(), s.cfg.IncludeTemplates)
	s.lastRefresh = time.Now()
}

// Entries returns a snapshot of every visible entry.
func (s *Service) Entries() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]Entry(nil), s.index...)
}

// List returns the direct children of a remote directory.
func (s *Service) List(dir string) ([]Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	node, err := s.resolveDirLocked(dir)
	if err != nil {
		return nil, err
	}
	return ChildrenEntries(node, s.cfg.IncludeTemplates), nil
}

// Search ranks the index against a query.
func (s *Service) Search(query string, limit int) []SearchResult {
	s.mu.Lock()
	idx := append([]Entry(nil), s.index...)
	s.mu.Unlock()
	return Search(idx, query, limit)
}

// Status summarizes service state for the UI.
func (s *Service) Status() Status {
	s.mu.Lock()
	count := len(s.index)
	last := s.lastRefresh
	s.mu.Unlock()
	_, pandocErr := exec.LookPath("pandoc")
	return Status{
		Authenticated:    s.apiCtx != nil,
		DocumentCount:    count,
		LastRefresh:      last,
		DefaultRemoteDir: s.cfg.DefaultRemoteDir,
		PandocAvailable:  pandocErr == nil,
	}
}

// Refresh forces a remote resync (mirror the hash tree) and rebuilds the index.
// The lock is held for the duration, so reads block until the resync finishes;
// this is acceptable for an explicit user action.
func (s *Service) Refresh(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.apiCtx == nil {
		return ErrInvalid
	}
	if _, _, err := s.apiCtx.Refresh(); err != nil {
		return errors.Wrap(err, "refresh cloud tree")
	}
	s.reindexLocked()
	return nil
}

// resolveDirLocked resolves a path to a directory node. Callers must hold s.mu.
func (s *Service) resolveDirLocked(dir string) (*model.Node, error) {
	if s.apiCtx == nil {
		return nil, ErrInvalid
	}
	node, err := s.apiCtx.Filetree().NodeByPath(normalizePath(dir), nil)
	if err != nil || node == nil {
		return nil, ErrNotFound
	}
	if node.IsFile() {
		return nil, ErrNotDir
	}
	return node, nil
}

// nodeByIDLocked resolves a document ID to a node. Callers must hold s.mu.
func (s *Service) nodeByIDLocked(id string) (*model.Node, error) {
	if s.apiCtx == nil {
		return nil, ErrInvalid
	}
	node := s.apiCtx.Filetree().NodeById(id)
	if node == nil || node.IsRoot() {
		return nil, ErrNotFound
	}
	return node, nil
}

// mutate runs fn under the service lock, syncs the cloud tree, and reindexes.
func (s *Service) mutate(fn func(api.ApiCtx) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.apiCtx == nil {
		return ErrInvalid
	}
	if err := fn(s.apiCtx); err != nil {
		return err
	}
	if err := s.apiCtx.SyncComplete(); err != nil {
		return errors.Wrap(err, "sync cloud tree")
	}
	s.reindexLocked()
	return nil
}
