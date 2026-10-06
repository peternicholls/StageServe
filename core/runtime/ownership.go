package runtime

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Ownership carries the durable ledger and the journaled mutation identity.
type Ownership struct {
	InstallationID string
	ProjectID      string
	OperationID    string
	Scope          string
	Resources      []Resource
}

// Resource identifies one incarnation. CreationOperation distinguishes reuse of
// Apple's name-based IDs after deletion; a tombstone must never be resurrected.
type Resource struct {
	Kind              string
	ID                string
	Name              string
	InstallationID    string
	ProjectID         string
	Role              string
	CreationOperation string
	Planned           bool
	Deleted           bool
}

type ResourceRecorder func(context.Context, Resource) error

var ownershipUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func (o Ownership) Validate() error {
	if !ownershipUUID.MatchString(o.InstallationID) || !ownershipUUID.MatchString(o.OperationID) {
		return fmt.Errorf("runtime ownership requires installation and operation UUIDs")
	}
	if o.Scope != "project" && o.Scope != "installation" {
		return fmt.Errorf("runtime ownership scope must be project or installation")
	}
	if !ownershipUUID.MatchString(o.ProjectID) || (o.Scope == "project" && o.ProjectID == o.InstallationID) || (o.Scope == "installation" && o.ProjectID != o.InstallationID) {
		return fmt.Errorf("runtime ownership project UUID does not match scope")
	}

	seen := map[string]bool{}
	live := map[string]bool{}
	for _, r := range o.Resources {
		if strings.TrimSpace(r.Kind) == "" || strings.TrimSpace(r.ID) == "" || strings.TrimSpace(r.Name) == "" || strings.TrimSpace(r.Role) == "" || r.InstallationID != o.InstallationID || r.ProjectID != o.ProjectID || !ownershipUUID.MatchString(r.CreationOperation) || (r.Planned && r.Deleted) {
			return fmt.Errorf("runtime ledger has invalid resource ownership or provenance")
		}
		key := r.Kind + "\x00" + r.ID
		incarnation := key + "\x00" + r.CreationOperation
		if seen[incarnation] || (!r.Deleted && live[key]) {
			return fmt.Errorf("runtime ledger has duplicate resource incarnations")
		}
		seen[incarnation] = true
		if !r.Deleted {
			live[key] = true
		}
	}
	return nil
}
