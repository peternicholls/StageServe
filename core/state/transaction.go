package state

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
)

// LifecycleStore persists the runtime result through a recoverable multi-file commit.
type LifecycleStore interface {
	StateStore
	IdentityStore
	CommitOperation(Operation, Identity, *Record, uint64) error
	InstallationIdentity(string) (Identity, error)
	CommitInstallationOperation(Operation, Identity, uint64) error
	RecoverOperation(string) error
}

// CommitPayload is durable recovery intent. PriorRecord preserves exact bytes so
// recovery can distinguish its own projection from a conflicting external write.
type CommitPayload struct {
	Installation       bool     `json:"installation,omitempty"`
	PriorIdentity      Identity `json:"prior_identity"`
	TargetIdentity     Identity `json:"target_identity"`
	PriorRecord        []byte   `json:"prior_record,omitempty"`
	PriorRecordPresent bool     `json:"prior_record_present"`
	Record             *Record  `json:"record,omitempty"`
}

func equalCommit(a, b *CommitPayload) bool { return reflect.DeepEqual(a, b) }
func validateCommitPayload(op Operation) error {
	p := op.Commit
	if p == nil {
		if op.Phase == "committing" {
			return ErrInvalidIdentity
		}
		return nil
	}
	if op.Phase != "committing" && op.Phase != "committed" {
		return ErrInvalidIdentity
	}
	before, after := p.PriorIdentity, p.TargetIdentity
	if before.ProjectID != op.ProjectID || before.InstallationID != op.InstallationID || after.ProjectID != before.ProjectID || after.InstallationID != before.InstallationID || after.CanonicalPath != before.CanonicalPath || after.Slug != before.Slug {
		return ErrOwnerMismatch
	}
	if p.Installation != (op.ProjectID == op.InstallationID) {
		return ErrOwnerMismatch
	}
	if p.Installation && (p.Record != nil || p.PriorRecordPresent || len(p.PriorRecord) != 0 || !before.Registered || !after.Registered || before.Slug != "installation") {
		return ErrOwnerMismatch
	}
	if before.Revision == 0 || before.Revision == ^uint64(0) || after.Revision != before.Revision+1 || op.BaseRevision > before.Revision {
		return ErrRevision
	}
	if err := validateIdentity(before, op.InstallationID); err != nil {
		return err
	}
	if err := validateIdentity(after, op.InstallationID); err != nil {
		return err
	}
	if err := preserveResourceIdentity(before.Resources, after.Resources); err != nil {
		return err
	}
	for _, r := range before.Resources {
		found := false
		for _, next := range after.Resources {
			if sameIncarnation(r, next) {
				found = true
			}
		}
		if !found {
			return ErrOwnerMismatch
		}
	}
	if p.PriorRecordPresent {
		var rec Record
		if err := json.Unmarshal(p.PriorRecord, &rec); err != nil {
			return ErrInvalidIdentity
		}
		if err := recordMatchesIdentity(rec, before); err != nil {
			return err
		}
	} else if len(p.PriorRecord) != 0 {
		return ErrInvalidIdentity
	}
	// Every resource observed by this operation remains independently recoverable,
	// including deleted tombstones; a projection cannot erase journal provenance.
	for _, resource := range append(append([]ResourceRecord{}, op.Resources...), op.Leftovers...) {
		found := false
		for _, retained := range after.Resources {
			if reflect.DeepEqual(resource, retained) {
				found = true
				break
			}
		}
		if !found {
			return ErrOwnerMismatch
		}
	}
	if p.Installation {
		return nil
	}
	if p.Record == nil {
		if after.Registered {
			return ErrOwnerMismatch
		}
	} else {
		if !after.Registered {
			return ErrOwnerMismatch
		}
		if err := recordMatchesIdentity(*p.Record, after); err != nil {
			return err
		}
	}
	return nil
}
func recordMatchesIdentity(rec Record, id Identity) error {
	if rec.SchemaVersion != SchemaVersion {
		return ErrSchema
	}
	if err := normalizeRecord(&rec); err != nil {
		return err
	}
	if rec.ProjectID != id.ProjectID || rec.InstallationID != id.InstallationID || rec.Project.Slug != id.Slug {
		return ErrOwnerMismatch
	}
	path, err := canonical(rec.Project.Dir)
	if err != nil {
		return err
	}
	if path != id.CanonicalPath {
		return ErrOwnerMismatch
	}
	return nil
}
func (s *Store) rejectCommitting(id string) error {
	ops, err := s.pending(id)
	if err != nil {
		return err
	}
	for _, op := range ops {
		if op.Phase == "committing" {
			return ErrRevision
		}
	}
	return nil
}
func (s *Store) transactionBoundary(name string) error {
	if s.transactionFault != nil {
		return s.transactionFault(name)
	}
	return nil
}

func (s *Store) CommitOperation(op Operation, updated Identity, rec *Record, revision uint64) error {
	return s.commitOperation(op, updated, rec, revision, false)
}
func (s *Store) CommitInstallationOperation(op Operation, updated Identity, revision uint64) error {
	return s.commitOperation(op, updated, nil, revision, true)
}
func (s *Store) commitOperation(op Operation, updated Identity, rec *Record, revision uint64, installation bool) error {
	if installation != (op.ProjectID == op.InstallationID) {
		return ErrOwnerMismatch
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return err
	}
	defer unlock()
	old, err := s.operation(op.ID)
	if err != nil {
		return err
	}
	if old.Commit != nil {
		target := updated
		target.Revision = revision + 1
		var normalized *Record
		if rec != nil {
			copy := *rec
			copy.SchemaVersion = SchemaVersion
			if err := normalizeRecord(&copy); err != nil {
				return err
			}
			normalized = &copy
		}
		if old.Commit.Installation != installation {
			return ErrOwnerMismatch
		}
		if !reflect.DeepEqual(old.Commit.TargetIdentity, target) || !reflect.DeepEqual(old.Commit.Record, normalized) || old.Commit.PriorIdentity.Revision != revision {
			return ErrRevision
		}
		return s.finishCommit(old)
	}
	if op.Commit != nil || !reflect.DeepEqual(old, op) || old.Phase == "committed" || old.Phase == "rolled_back" || old.Phase == "rollback" || old.Phase == "degraded" {
		return ErrRevision
	}
	current, err := s.identity(op.ProjectID)
	if err != nil {
		return err
	}
	if current.Revision != revision || updated.Revision != revision {
		return ErrRevision
	}
	if revision == ^uint64(0) {
		return ErrRevision
	}
	updated.Revision++
	var normalized *Record
	if rec != nil {
		copy := *rec
		if copy.SchemaVersion != 0 && copy.SchemaVersion != SchemaVersion {
			return ErrSchema
		}
		copy.SchemaVersion = SchemaVersion
		if err := normalizeRecord(&copy); err != nil {
			return err
		}
		normalized = &copy
	}
	var prior []byte
	present := false
	if !installation {
		prior, err = os.ReadFile(s.projectFile(current.Slug))
		present = err == nil
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	op.Commit = &CommitPayload{Installation: installation, PriorIdentity: current, TargetIdentity: updated, PriorRecord: prior, PriorRecordPresent: present, Record: normalized}
	op.Phase = "committing"
	if err := validateCommitPayload(op); err != nil {
		return err
	}
	terminal := op
	terminal.Phase = "committed"
	if err := validateTerminalOperation(terminal); err != nil {
		return err
	}
	if err := durableJSON(filepath.Join(s.stateDir, "operations", op.ID+".json"), op); err != nil {
		return err
	}
	if err := s.transactionBoundary("journal"); err != nil {
		return err
	}
	return s.finishCommit(op)
}
func (s *Store) finishCommit(op Operation) error {
	if err := validateCommitPayload(op); err != nil {
		return err
	}
	p := op.Commit
	if p == nil {
		return ErrInvalidIdentity
	}
	terminal := op
	terminal.Phase = "committed"
	if err := validateTerminalOperation(terminal); err != nil {
		return err
	}
	current, err := s.identity(op.ProjectID)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(current, p.PriorIdentity) && !reflect.DeepEqual(current, p.TargetIdentity) {
		return ErrRevision
	}
	var path string
	targetMatch := true
	if !p.Installation {
		path = s.projectFile(p.TargetIdentity.Slug)
		actual, err := os.ReadFile(path)
		present := err == nil
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		var target []byte
		if p.Record != nil {
			target, err = json.MarshalIndent(p.Record, "", "  ")
			if err != nil {
				return err
			}
		}
		priorMatch := present == p.PriorRecordPresent && bytes.Equal(actual, p.PriorRecord)
		targetMatch = present == (p.Record != nil) && bytes.Equal(actual, target)
		if !priorMatch && !targetMatch {
			return ErrOwnerMismatch
		}
	}
	if op.Phase == "committed" {
		if !reflect.DeepEqual(current, p.TargetIdentity) || !targetMatch {
			return ErrRevision
		}
		return nil
	}
	if err := durableJSON(s.identityFile(op.ProjectID, op.InstallationID), p.TargetIdentity); err != nil {
		return err
	}
	if err := s.transactionBoundary("identity"); err != nil {
		return err
	}
	if !p.Installation {
		if p.Record != nil {
			err = durableJSON(path, p.Record)
		} else {
			err = durableRemove(path)
		}
		if err != nil {
			return err
		}
		if err := s.transactionBoundary("record"); err != nil {
			return err
		}
	}
	op.Phase = "committed"
	if err := durableJSON(filepath.Join(s.stateDir, "operations", op.ID+".json"), op); err != nil {
		return err
	}
	return s.transactionBoundary("committed")
}
func durableRemove(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	d, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

// RecoverOperation replays durable committing intent only. Earlier runtime
// phases require explicit reconciliation; they are never reported as recovered.
func (s *Store) RecoverOperation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	unlock, err := s.lockIdentityState()
	if err != nil {
		return err
	}
	defer unlock()
	ops, err := s.pending(id)
	if err != nil {
		return err
	}
	if len(ops) > 1 {
		return ErrCollision
	}
	var committing *Operation
	for i := range ops {
		if ops[i].Phase == "committing" {
			if committing != nil {
				return ErrCollision
			}
			committing = &ops[i]
		}
	}
	if committing == nil {
		if len(ops) != 0 {
			return ErrCollision
		}
		return nil
	}
	return s.finishCommit(*committing)
}

var _ LifecycleStore = (*Store)(nil)
