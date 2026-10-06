// State store: per-project JSON files, atomic writes, registry projection.
//
// Atomicity is enforced via temp-file + os.Rename (FR-008). Concurrent access
// across Store instances and processes is serialised with the identity lock.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/peternicholls/stageserve/core/runtime"
)

const SchemaVersion = 2

// ErrNotFound is returned when a slug has no recorded state.
var ErrNotFound = errors.New("state: project not found")

// Store is the default StateStore implementation.
type Store struct {
	stateDir         string
	mu               sync.Mutex
	transactionFault func(string) error
}

// NewStore returns a Store rooted at stateDir. It ensures stateDir/projects exists.
func NewStore(stateDir string) (*Store, error) {
	if stateDir == "" {
		return nil, errors.New("state: empty state dir")
	}
	if err := os.MkdirAll(filepath.Join(stateDir, "projects"), 0o755); err != nil {
		return nil, err
	}
	return &Store{stateDir: stateDir}, nil
}

// StateDir returns the configured state directory.
func (s *Store) StateDir() string { return s.stateDir }

func (s *Store) projectFile(slug string) string {
	return filepath.Join(s.stateDir, "projects", slug+".json")
}

// Save writes a project record to disk atomically.
func (s *Store) Save(rec Record) error {
	if !validName(rec.Project.Slug) {
		return errors.New("state: cannot save record with empty slug")
	}
	if rec.SchemaVersion != 0 && rec.SchemaVersion != SchemaVersion {
		return ErrSchema
	}
	if err := normalizeRecord(&rec); err != nil {
		return err
	}
	rec.SchemaVersion = SchemaVersion
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return err
	}
	defer unlock()

	if err := os.MkdirAll(filepath.Join(s.stateDir, "projects"), 0o755); err != nil {
		return err
	}
	target := s.projectFile(rec.Project.Slug)

	old, err := s.loadRecordFile(target, false)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if err == nil && (old.ProjectID != rec.ProjectID || old.InstallationID != rec.InstallationID) {
		return ErrOwnerMismatch
	}
	if err := s.validateRecordIdentity(rec, true); err != nil {
		return err
	}
	return durableJSON(target, rec)
}

// Load reads the recorded state for slug.
func (s *Store) Load(slug string) (Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return Record{}, err
	}
	defer unlock()
	if !validName(slug) {
		return Record{}, ErrInvalidIdentity
	}
	return s.loadFile(s.projectFile(slug))
}

func (s *Store) loadFile(path string) (Record, error) {
	return s.loadRecordFile(path, true)
}

// Callers hold both the Store mutex and the cross-process identity lock.
func (s *Store) loadRecordFile(path string, registered bool) (Record, error) {
	var rec Record
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return rec, ErrNotFound
		}
		return rec, err
	}
	if err := json.Unmarshal(data, &rec); err != nil {
		return rec, fmt.Errorf("state: parse %s: %w", path, err)
	}
	if rec.SchemaVersion != SchemaVersion {
		return rec, ErrSchema
	}
	if err := normalizeRecord(&rec); err != nil {
		return rec, fmt.Errorf("state: parse %s: %w", path, err)
	}
	if rec.Project.Slug != strings.TrimSuffix(filepath.Base(path), ".json") {
		return rec, ErrOwnerMismatch
	}
	if err := s.validateRecordIdentity(rec, registered); err != nil {
		return rec, err
	}
	return rec, nil
}

// UUID-free records remain readable without creating or adopting ownership.
func (s *Store) validateRecordIdentity(rec Record, registered bool) error {
	if rec.ProjectID == "" && rec.InstallationID == "" {
		return nil
	}
	if !validUUID(rec.ProjectID) || !validUUID(rec.InstallationID) {
		return ErrInvalidIdentity
	}
	if rec.ProjectID == rec.InstallationID {
		return ErrOwnerMismatch
	}
	identity, err := s.identity(rec.ProjectID)
	if errors.Is(err, ErrNotFound) {
		return ErrOwnerMismatch
	}
	if err != nil {
		return err
	}
	if identity.InstallationID != rec.InstallationID || identity.Slug != rec.Project.Slug || (registered && !identity.Registered) {
		return ErrOwnerMismatch
	}
	path, err := canonical(rec.Project.Dir)
	if err != nil {
		return err
	}
	if path != identity.CanonicalPath {
		return ErrOwnerMismatch
	}
	return nil
}

func normalizeRecord(rec *Record) error {
	if rec.Project.RuntimeBackend == "" || rec.Project.RuntimeBackend == "docker" {
		return ErrLegacyState
	}
	backend, err := runtime.ParseBackend(string(rec.Project.RuntimeBackend))
	if err != nil {
		return err
	}
	rec.Project.RuntimeBackend = backend
	if rec.Runtime.Backend == "" {
		rec.Runtime.Backend = backend
	} else if _, err := runtime.ParseBackend(string(rec.Runtime.Backend)); err != nil {
		return err
	}
	if rec.Runtime.Backend != backend {
		return ErrOwnerMismatch
	}
	return nil
}

// Remove deletes the state record for slug. Idempotent.
func (s *Store) Remove(slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return err
	}
	defer unlock()
	if !validName(slug) {
		return ErrInvalidIdentity
	}
	if _, err := s.loadRecordFile(s.projectFile(slug), false); err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	err = os.Remove(s.projectFile(slug))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// listFiles returns every per-project state file (sorted for determinism).
// Includes .json files only.
func (s *Store) listFiles() ([]string, error) {
	dir := filepath.Join(s.stateDir, "projects")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".json") {
			out = append(out, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(out)
	return out, nil
}

// StateFileForSelector resolves a selector against the recorded projects.
// Selectors match against slug, name, hostname, or project dir (mirrors
// stageserve_state_file_for_selector).
func (s *Store) StateFileForSelector(selector string) (Record, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return Record{}, "", err
	}
	defer unlock()

	files, err := s.listFiles()
	if err != nil {
		return Record{}, "", err
	}
	for _, f := range files {
		rec, err := s.loadFile(f)
		if err != nil {
			return Record{}, "", err
		}
		p := rec.Project
		if selector == p.Slug || selector == p.Name || selector == p.Hostname || selector == p.Dir {
			return rec, f, nil
		}
	}
	return Record{}, "", ErrNotFound
}
