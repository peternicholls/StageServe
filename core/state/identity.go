package state

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrSchema          = errors.New("state: unsupported schema; restore a compatible backup")
	ErrLegacyState     = errors.New("state: legacy runtime identity; explicit export/import required")
	ErrOwnerMismatch   = errors.New("state: resource ownership mismatch")
	ErrInvalidIdentity = errors.New("state: invalid identity")
	ErrRevision        = errors.New("state: stale revision")
	ErrCollision       = errors.New("state: identity collision; explicit recovery required")
)

const IdentitySchemaVersion = 1

type ResourceRecord struct {
	Kind              string `json:"kind"`
	ID                string `json:"id"`
	Name              string `json:"name"`
	InstallationID    string `json:"installation_id"`
	ProjectID         string `json:"project_id"`
	Role              string `json:"role"`
	CreationOperation string `json:"creation_operation"`
	Deleted           bool   `json:"deleted,omitempty"`
	Planned           bool   `json:"planned,omitempty"`
}

type Identity struct {
	SchemaVersion  int              `json:"schema_version"`
	InstallationID string           `json:"installation_id"`
	ProjectID      string           `json:"project_id"`
	CanonicalPath  string           `json:"canonical_path"`
	Slug           string           `json:"slug"`
	Registered     bool             `json:"registered"`
	Revision       uint64           `json:"revision"`
	Resources      []ResourceRecord `json:"resources,omitempty"`
	DesiredHash    string           `json:"desired_hash,omitempty"`
	AppliedHash    string           `json:"applied_hash,omitempty"`
	EngineDigest   string           `json:"engine_digest,omitempty"`
	EngineVersion  string           `json:"engine_version,omitempty"`
	LastWriter     string           `json:"last_writer,omitempty"`
}

type Operation struct {
	Commit             *CommitPayload   `json:"commit,omitempty"`
	SchemaVersion      int              `json:"schema_version"`
	ID                 string           `json:"id"`
	InstallationID     string           `json:"installation_id"`
	ProjectID          string           `json:"project_id"`
	BaseRevision       uint64           `json:"base_revision"`
	Kind               string           `json:"kind"`
	DesiredHash        string           `json:"desired_hash,omitempty"`
	Phase              string           `json:"phase"`
	Resources          []ResourceRecord `json:"resources,omitempty"`
	Leftovers          []ResourceRecord `json:"leftovers,omitempty"`
	PriorConfiguration json.RawMessage  `json:"prior_configuration,omitempty"`
}

// IdentityStore is separate from legacy registry storage so callers explicitly
// require ownership persistence before enabling runtime mutations.
type IdentityStore interface {
	CreateIdentity(path, slug string) (Identity, error)
	LoadIdentity(projectID string) (Identity, error)
	IdentityForPath(path string) (Identity, error)
	SaveIdentity(identity Identity, expectedRevision uint64) error
	BeginOperation(projectID string, baseRevision uint64, kind, desiredHash string) (Operation, error)
	SaveOperation(operation Operation) error
	PendingOperations(projectID string) ([]Operation, error)
}

func validName(s string) bool {
	return s != "" && s != "." && s != ".." && !strings.ContainsAny(s, "/\\\x00")
}
func validUUID(s string) bool {
	if len(s) != 36 || s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	_, e := hex.DecodeString(strings.ReplaceAll(s, "-", ""))
	return e == nil
}
func newUUID() (string, error) {
	var b [16]byte
	if _, e := rand.Read(b[:]); e != nil {
		return "", e
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	h := hex.EncodeToString(b[:])
	return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:], nil
}
func canonical(path string) (string, error) {
	p, e := filepath.Abs(path)
	if e != nil {
		return "", e
	}
	return filepath.EvalSymlinks(p)
}

// durableJSON syncs both contents and the directory entry; temporary files are
// private and removed on every error. Callers serialize compound operations.
func durableJSON(path string, value any) error {
	if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	b, e := json.MarshalIndent(value, "", "  ")
	if e != nil {
		return e
	}
	f, e := os.CreateTemp(filepath.Dir(path), ".state-*")
	if e != nil {
		return e
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if _, e = f.Write(b); e != nil {
		return e
	}
	if e = f.Sync(); e != nil {
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	if e = os.Rename(f.Name(), path); e != nil {
		return e
	}
	d, e := os.Open(filepath.Dir(path))
	if e != nil {
		return e
	}
	defer d.Close()
	return d.Sync()
}
func readJSON(path string, v any) error {
	b, e := os.ReadFile(path)
	if os.IsNotExist(e) {
		return ErrNotFound
	}
	if e != nil {
		return e
	}
	if e = json.Unmarshal(b, v); e != nil {
		return fmt.Errorf("state: corrupt %s: %w", path, e)
	}
	return nil
}
func (s *Store) installation() (string, error) {
	var v struct {
		SchemaVersion int    `json:"schema_version"`
		ID            string `json:"id"`
	}
	e := readJSON(filepath.Join(s.stateDir, "installation.json"), &v)
	if e != nil {
		return "", e
	}
	if v.SchemaVersion != 1 {
		return "", ErrSchema
	}
	if !validUUID(v.ID) {
		return "", ErrInvalidIdentity
	}
	return v.ID, nil
}
func (s *Store) initInstallation() (string, error) {
	id, e := s.installation()
	if !errors.Is(e, ErrNotFound) {
		return id, e
	}
	if _, err := os.Stat(filepath.Join(s.stateDir, "installation-resources.json")); err == nil {
		return "", ErrOwnerMismatch
	} else if !os.IsNotExist(err) {
		return "", err
	}
	// Existing ownership without an installation record is damaged, never adopt it.
	for _, dir := range []string{"identities", "operations"} {
		entries, err := os.ReadDir(filepath.Join(s.stateDir, dir))
		if err != nil && !os.IsNotExist(err) {
			return "", err
		}
		if len(entries) > 0 {
			return "", ErrOwnerMismatch
		}
	}
	id, e = newUUID()
	if e != nil {
		return "", e
	}
	e = durableJSON(filepath.Join(s.stateDir, "installation.json"), struct {
		SchemaVersion int    `json:"schema_version"`
		ID            string `json:"id"`
	}{1, id})
	return id, e
}
func (s *Store) identityFile(id, owner string) string {
	if id == owner {
		return filepath.Join(s.stateDir, "installation-resources.json")
	}
	return filepath.Join(s.stateDir, "identities", id+".json")
}
func (s *Store) identity(id string) (Identity, error) {
	var v Identity
	if !validUUID(id) {
		return v, ErrInvalidIdentity
	}
	owner, err := s.installation()
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			if _, e := os.Stat(filepath.Join(s.stateDir, "installation-resources.json")); e == nil {
				return v, ErrOwnerMismatch
			} else if !os.IsNotExist(e) {
				return v, e
			}
			if _, e := os.Stat(filepath.Join(s.stateDir, "identities", id+".json")); e == nil {
				return v, ErrOwnerMismatch
			} else if !os.IsNotExist(e) {
				return v, e
			}
		}
		return v, err
	}
	if err := readJSON(s.identityFile(id, owner), &v); err != nil {
		return v, err
	}
	if err := validateIdentity(v, owner); err != nil {
		return v, err
	}
	if v.ProjectID != id {
		return v, ErrOwnerMismatch
	}
	if id == owner && (!v.Registered || v.Slug != "installation") {
		return v, ErrOwnerMismatch
	}
	return v, nil
}

// InstallationIdentity reserves the installation UUID exclusively for shared
// resources. It is persisted separately and never appears in the project registry.
func (s *Store) InstallationIdentity(path string) (Identity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return Identity{}, err
	}
	defer unlock()
	canonicalPath, err := canonical(path)
	if err != nil {
		return Identity{}, err
	}
	owner, err := s.initInstallation()
	if err != nil {
		return Identity{}, err
	}
	id, err := s.identity(owner)
	if err == nil {
		if id.CanonicalPath != canonicalPath {
			return Identity{}, ErrOwnerMismatch
		}
		return id, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return Identity{}, err
	}
	entries, err := os.ReadDir(filepath.Join(s.stateDir, "operations"))
	if err != nil && !os.IsNotExist(err) {
		return Identity{}, err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		var op Operation
		if err := readJSON(filepath.Join(s.stateDir, "operations", entry.Name()), &op); err != nil {
			return Identity{}, err
		}
		if op.ProjectID == owner {
			return Identity{}, ErrOwnerMismatch
		}
	}
	id = Identity{SchemaVersion: IdentitySchemaVersion, InstallationID: owner, ProjectID: owner, CanonicalPath: canonicalPath, Slug: "installation", Registered: true, Revision: 1}
	err = durableJSON(s.identityFile(owner, owner), id)
	return id, err
}

func validateResources(rs []ResourceRecord, owner, id string) error {
	seen := map[string]bool{}
	live := map[string]bool{}
	for _, r := range rs {
		if r.InstallationID != owner || r.ProjectID != id {
			return ErrOwnerMismatch
		}
		if r.Planned && r.Deleted {
			return ErrInvalidIdentity
		}
		if r.Kind == "" || r.ID == "" || r.Role == "" || !validUUID(r.CreationOperation) {
			return ErrInvalidIdentity
		}
		k := r.Kind + ":" + r.ID + ":" + r.CreationOperation
		if !r.Deleted {
			name := r.Kind + ":" + r.ID
			if live[name] {
				return ErrCollision
			}
			live[name] = true
		}
		if seen[k] {
			return ErrCollision
		}
		seen[k] = true
	}
	return nil
}
func validateIdentity(v Identity, owner string) error {
	if v.SchemaVersion != IdentitySchemaVersion {
		return ErrSchema
	}
	if v.InstallationID != owner {
		return ErrOwnerMismatch
	}
	if !validUUID(v.ProjectID) || !filepath.IsAbs(v.CanonicalPath) || !validName(v.Slug) || v.Revision == 0 {
		return ErrInvalidIdentity
	}
	for _, resource := range v.Resources {
		if resource.Planned {
			return ErrInvalidIdentity
		}
	}
	return validateResources(v.Resources, owner, v.ProjectID)
}
func (s *Store) identities() ([]Identity, error) {
	// Even an empty ledger cannot hide damaged installation metadata. Missing
	// metadata is a new installation only when no ownership/journals remain.
	_, installationErr := s.installation()
	if installationErr != nil {
		if !errors.Is(installationErr, ErrNotFound) {
			return nil, installationErr
		}
		if _, err := os.Stat(filepath.Join(s.stateDir, "installation-resources.json")); err == nil {
			return nil, ErrOwnerMismatch
		} else if !os.IsNotExist(err) {
			return nil, err
		}
		for _, dir := range []string{"identities", "operations"} {
			entries, err := os.ReadDir(filepath.Join(s.stateDir, dir))
			if err != nil && !os.IsNotExist(err) {
				return nil, err
			}
			if len(entries) > 0 {
				return nil, ErrOwnerMismatch
			}
		}
	}
	entries, e := os.ReadDir(filepath.Join(s.stateDir, "identities"))
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var out []Identity
	paths := map[string]bool{}
	for _, f := range entries {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		owner, err := s.installation()
		if err != nil {
			return nil, err
		}
		if strings.TrimSuffix(f.Name(), ".json") == owner {
			return nil, ErrOwnerMismatch
		}
		v, e := s.identity(strings.TrimSuffix(f.Name(), ".json"))
		if e != nil {
			return nil, e
		}
		if paths[v.CanonicalPath] {
			return nil, ErrCollision
		}
		paths[v.CanonicalPath] = true
		out = append(out, v)
	}
	return out, nil
}
func (s *Store) LoadIdentity(id string) (Identity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return Identity{}, err
	}
	defer unlock()
	return s.identity(id)
}
func (s *Store) IdentityForPath(path string) (Identity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return Identity{}, err
	}
	defer unlock()
	p, e := canonical(path)
	if e != nil {
		return Identity{}, e
	}
	vs, e := s.identities()
	if e != nil {
		return Identity{}, e
	}
	for _, v := range vs {
		if v.CanonicalPath == p {
			return v, nil
		}
	}
	return Identity{}, ErrNotFound
}
func (s *Store) CreateIdentity(path, slug string) (Identity, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return Identity{}, err
	}
	defer unlock()
	if !validName(slug) {
		return Identity{}, ErrInvalidIdentity
	}
	p, e := canonical(path)
	if e != nil {
		return Identity{}, e
	}
	vs, e := s.identities()
	if e != nil {
		return Identity{}, e
	}
	for _, v := range vs {
		if v.CanonicalPath == p {
			if v.Slug != slug {
				return Identity{}, ErrCollision
			}
			return v, nil
		}
		if v.Slug == slug {
			return Identity{}, ErrCollision
		}
	}
	owner, e := s.initInstallation()
	if e != nil {
		return Identity{}, e
	}
	id, e := newUUID()
	if e != nil {
		return Identity{}, e
	}
	v := Identity{SchemaVersion: 1, InstallationID: owner, ProjectID: id, CanonicalPath: p, Slug: slug, Registered: true, Revision: 1}
	if id == owner {
		return Identity{}, ErrCollision
	}
	e = durableJSON(filepath.Join(s.stateDir, "identities", id+".json"), v)
	return v, e
}
func (s *Store) SaveIdentity(v Identity, revision uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return err
	}
	defer unlock()
	old, e := s.identity(v.ProjectID)
	if e != nil {
		return e
	}
	if err := s.rejectCommitting(v.ProjectID); err != nil {
		return err
	}
	if old.Revision != revision || v.Revision != revision {
		return ErrRevision
	}
	if v.InstallationID != old.InstallationID || v.CanonicalPath != old.CanonicalPath {
		return ErrOwnerMismatch
	}
	if e = validateIdentity(v, old.InstallationID); e != nil {
		return e
	}
	vs, e := s.identities()
	if e != nil {
		return e
	}
	for _, other := range vs {
		if v.ProjectID != v.InstallationID && other.ProjectID != v.ProjectID && other.Slug == v.Slug {
			return ErrCollision
		}
	}
	if e = preserveResourceIdentity(old.Resources, v.Resources); e != nil {
		return e
	}
	for _, before := range old.Resources {
		retained := false
		for _, after := range v.Resources {
			if sameIncarnation(before, after) {
				retained = true
				break
			}
		}
		if !retained {
			return ErrOwnerMismatch
		}
	}
	if v.Revision == ^uint64(0) {
		return ErrRevision
	}
	v.Revision++
	if v.ProjectID == v.InstallationID && (!v.Registered || v.Slug != "installation") {
		return ErrOwnerMismatch
	}
	return durableJSON(s.identityFile(v.ProjectID, v.InstallationID), v)
}
func (s *Store) BeginOperation(id string, revision uint64, kind, hash string) (Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return Operation{}, err
	}
	defer unlock()
	v, e := s.identity(id)
	if e != nil {
		return Operation{}, e
	}
	if v.Revision != revision {
		return Operation{}, ErrRevision
	}
	if !validName(kind) {
		return Operation{}, ErrInvalidIdentity
	}
	pending, e := s.pending(id)
	if e != nil {
		return Operation{}, e
	}
	if len(pending) > 0 {
		return Operation{}, ErrCollision
	}
	opID, e := newUUID()
	if e != nil {
		return Operation{}, e
	}
	op := Operation{SchemaVersion: 1, ID: opID, InstallationID: v.InstallationID, ProjectID: id, BaseRevision: revision, Kind: kind, DesiredHash: hash, Phase: "intent"}
	e = durableJSON(filepath.Join(s.stateDir, "operations", opID+".json"), op)
	return op, e
}
func (s *Store) operation(id string) (Operation, error) {
	var op Operation
	if !validUUID(id) {
		return op, ErrInvalidIdentity
	}
	e := readJSON(filepath.Join(s.stateDir, "operations", id+".json"), &op)
	if e != nil {
		return op, e
	}
	if op.SchemaVersion != 1 {
		return op, ErrSchema
	}
	v, e := s.identity(op.ProjectID)
	if e != nil {
		return op, e
	}
	if op.ID != id || op.InstallationID != v.InstallationID {
		return op, ErrOwnerMismatch
	}
	if !validName(op.Kind) || !validOperationPhase(op.Phase) || op.BaseRevision == 0 || op.BaseRevision > v.Revision {
		return op, ErrInvalidIdentity
	}
	if e = validateTerminalOperation(op); e != nil {
		return op, e
	}
	if e = validateCommitPayload(op); e != nil {
		return op, e
	}
	return op, validateResources(append(append([]ResourceRecord{}, op.Resources...), op.Leftovers...), op.InstallationID, op.ProjectID)
}
func (s *Store) SaveOperation(op Operation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return err
	}
	defer unlock()
	old, e := s.operation(op.ID)
	if e != nil {
		return e
	}
	if old.Phase == "committing" || op.Phase == "committing" || !equalCommit(old.Commit, op.Commit) {
		return ErrOwnerMismatch
	}
	if op.ProjectID != old.ProjectID || op.InstallationID != old.InstallationID || op.BaseRevision != old.BaseRevision || op.Kind != old.Kind || op.DesiredHash != old.DesiredHash {
		return ErrOwnerMismatch
	}
	if op.SchemaVersion != 1 {
		return ErrSchema
	}
	if !validOperationPhase(op.Phase) {
		return ErrInvalidIdentity
	}
	if !operationTransition(old.Phase, op.Phase) {
		return ErrRevision
	}
	if len(old.PriorConfiguration) > 0 && !sameJSON(old.PriorConfiguration, op.PriorConfiguration) {
		return ErrOwnerMismatch
	}
	if e = validateTerminalOperation(op); e != nil {
		return e
	}
	previous := append(append([]ResourceRecord{}, old.Resources...), old.Leftovers...)
	current := append(append([]ResourceRecord{}, op.Resources...), op.Leftovers...)
	if e = preserveResourceIdentity(previous, current); e != nil {
		return e
	}
	if e = validateResources(current, op.InstallationID, op.ProjectID); e != nil {
		return e
	}
	return durableJSON(filepath.Join(s.stateDir, "operations", op.ID+".json"), op)
}
func (s *Store) pending(id string) ([]Operation, error) {
	if _, e := s.identity(id); e != nil {
		return nil, e
	}
	entries, e := os.ReadDir(filepath.Join(s.stateDir, "operations"))
	if os.IsNotExist(e) {
		return nil, nil
	}
	if e != nil {
		return nil, e
	}
	var out []Operation
	for _, f := range entries {
		if !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		op, e := s.operation(strings.TrimSuffix(f.Name(), ".json"))
		if e != nil {
			return nil, e
		}
		if op.ProjectID == id && op.Phase != "committed" && op.Phase != "rolled_back" {
			out = append(out, op)
		}
	}
	return out, nil
}
func (s *Store) PendingOperations(id string) ([]Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return nil, err
	}
	defer unlock()
	return s.pending(id)
}

func validOperationPhase(phase string) bool {
	return operationRank(phase) >= 0
}

func operationRank(phase string) int {
	for i, p := range []string{"intent", "prepare", "apply", "probe", "route", "committing", "committed", "rollback", "degraded", "rolled_back"} {
		if phase == p {
			return i
		}
	}
	return -1
}

func operationTransition(old, next string) bool {
	if old == "committing" {
		return next == "committed"
	}
	if old == "committed" || old == "rolled_back" {
		return false
	}
	if old == "rollback" || old == "degraded" {
		return next == "rollback" || next == "degraded" || next == "rolled_back"
	}
	return operationRank(next) >= operationRank(old)
}

func sameIncarnation(a, b ResourceRecord) bool {
	return a.Kind == b.Kind && a.ID == b.ID && a.CreationOperation == b.CreationOperation
}

// Name-based resource IDs may be reused only after the prior incarnation is
// deleted. Its tombstone retains immutable provenance and cannot be resurrected.
func preserveResourceIdentity(old, next []ResourceRecord) error {
	for _, before := range old {
		deleted := before.Deleted
		retained := false
		for _, after := range next {
			if sameIncarnation(before, after) {
				retained = true
				if before.Name != after.Name || before.InstallationID != after.InstallationID || before.ProjectID != after.ProjectID || before.Role != after.Role || (before.Deleted && !after.Deleted) || (!before.Planned && after.Planned) {
					return ErrOwnerMismatch
				}
				deleted = after.Deleted
			}
		}
		if before.Planned && !retained {
			return ErrOwnerMismatch
		}
		for _, after := range next {
			if before.Kind == after.Kind && before.ID == after.ID && before.CreationOperation != after.CreationOperation && !deleted {
				return ErrOwnerMismatch
			}
		}
	}
	return nil
}

func sameJSON(a, b json.RawMessage) bool {
	var left, right bytes.Buffer
	return json.Compact(&left, a) == nil && json.Compact(&right, b) == nil && bytes.Equal(left.Bytes(), right.Bytes())
}

// Terminal journals cannot conceal cleanup that still needs recovery.
func validateTerminalOperation(op Operation) error {
	if op.Phase == "committed" || op.Phase == "rolled_back" {
		for _, resource := range append(append([]ResourceRecord{}, op.Resources...), op.Leftovers...) {
			if resource.Planned {
				return ErrInvalidIdentity
			}
		}
		for _, resource := range op.Leftovers {
			if !resource.Deleted {
				return ErrInvalidIdentity
			}
		}
	}
	return nil
}
