package lifecycle

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/peternicholls/stageserve/core/config"
	runtime "github.com/peternicholls/stageserve/core/runtime"
	"github.com/peternicholls/stageserve/core/state"
	"github.com/peternicholls/stageserve/infra/gateway"
)

type transactionKey bool

type lifecycleTransaction struct {
	store           state.LifecycleStore
	identity        state.Identity
	operation       state.Operation
	record          *state.Record
	installation    bool
	needsCleanup    bool
	mutationStarted bool
	finished        bool
	recordingFailed bool
}

func transactionFrom(ctx context.Context) *lifecycleTransaction {
	return ctx.Value(transactionKey(false)).(*lifecycleTransaction)
}
func installationTransactionFrom(ctx context.Context) *lifecycleTransaction {
	return ctx.Value(transactionKey(true)).(*lifecycleTransaction)
}
func resourceKey(kind, id, creation string) string { return kind + "\x00" + id + "\x00" + creation }
func (t *lifecycleTransaction) ownership() runtime.Ownership {
	projectID, scope := t.identity.ProjectID, "project"
	if t.installation {
		projectID = t.identity.InstallationID
		scope = "installation"
	}
	resources := make([]runtime.Resource, 0, len(t.identity.Resources))
	retained := append([]state.ResourceRecord(nil), t.identity.Resources...)
	for _, r := range t.operation.Resources {
		if r.Planned {
			retained = append(retained, r)
		}
	}
	for _, r := range retained {
		p := r.ProjectID
		if t.installation {
			p = t.identity.InstallationID
		}
		resources = append(resources, runtime.Resource{Kind: r.Kind, ID: r.ID, Name: r.Name, InstallationID: r.InstallationID, ProjectID: p, Role: r.Role, CreationOperation: r.CreationOperation, Deleted: r.Deleted, Planned: r.Planned})
	}
	return runtime.Ownership{InstallationID: t.identity.InstallationID, ProjectID: projectID, OperationID: t.operation.ID, Scope: scope, Resources: resources}
}
func (t *lifecycleTransaction) phase(phase string) error {
	t.operation.Phase = phase
	if err := t.store.SaveOperation(t.operation); err != nil {
		return err
	}
	return t.refresh()
}
func (t *lifecycleTransaction) refresh() error {
	ops, err := t.store.PendingOperations(t.identity.ProjectID)
	if err != nil {
		return err
	}
	for _, op := range ops {
		if op.ID == t.operation.ID {
			t.operation = op
			return nil
		}
	}
	return nil
}
func (t *lifecycleTransaction) recordResource(_ context.Context, r runtime.Resource) error {
	p := r.ProjectID

	if r.InstallationID != t.identity.InstallationID || p != t.identity.ProjectID {
		return state.ErrOwnerMismatch
	}
	next := state.ResourceRecord{Kind: r.Kind, ID: r.ID, Name: r.Name, InstallationID: r.InstallationID, ProjectID: p, Role: r.Role, CreationOperation: r.CreationOperation, Deleted: r.Deleted, Planned: r.Planned}
	merge := func(rs []state.ResourceRecord) ([]state.ResourceRecord, error) {
		for i, old := range rs {
			if resourceKey(old.Kind, old.ID, old.CreationOperation) == resourceKey(next.Kind, next.ID, next.CreationOperation) {
				if old.Name != next.Name || old.Role != next.Role || old.InstallationID != next.InstallationID || old.ProjectID != next.ProjectID || (old.Deleted && !next.Deleted) {
					return nil, state.ErrOwnerMismatch
				}
				rs = append([]state.ResourceRecord(nil), rs...)
				rs[i] = next
				return rs, nil
			}
		}
		return append(append([]state.ResourceRecord(nil), rs...), next), nil
	}
	observed, err := merge(t.operation.Resources)
	if err != nil {
		return err
	}
	retained := t.identity.Resources
	if !r.Planned {
		retained, err = merge(t.identity.Resources)
		if err != nil {
			return err
		}
	}
	candidate := t.operation
	candidate.Resources = observed
	t.operation = candidate
	t.identity.Resources = retained
	t.mutationStarted = true
	if err := t.store.SaveOperation(candidate); err != nil {
		t.recordingFailed = true
		return err
	}
	if err := t.refresh(); err != nil {
		t.recordingFailed = true
		return err
	}

	return nil
}
func (t *lifecycleTransaction) setRecord(rec state.Record) error {
	rec.ProjectID = t.identity.ProjectID
	rec.InstallationID = t.identity.InstallationID
	rec.SchemaVersion = state.SchemaVersion
	t.record = &rec
	t.identity.Registered = true
	return nil
}
func (t *lifecycleTransaction) detach() error {
	t.record = nil
	t.identity.Registered = false
	return nil
}
func (t *lifecycleTransaction) commit() error {
	t.identity.LastWriter = t.operation.ID
	if t.installation {
		if t.finished {
			return nil
		}
		if err := t.store.CommitInstallationOperation(t.operation, t.identity, t.identity.Revision); err != nil {
			return err
		}
		id, err := t.store.LoadIdentity(t.identity.ProjectID)
		if err != nil {
			return err
		}
		t.identity = id
		t.finished = true
		return nil
	}

	return t.store.CommitOperation(t.operation, t.identity, t.record, t.identity.Revision)
}
func (t *lifecycleTransaction) ensureActive() error {
	if !t.finished {
		return nil
	}
	op, err := t.store.BeginOperation(t.identity.ProjectID, t.identity.Revision, t.operation.Kind, t.operation.DesiredHash)
	if err != nil {
		return err
	}
	t.operation = op
	t.finished = false
	return t.phase("apply")
}
func (o *Orchestrator) beginTransaction(ctx context.Context, cfg config.ProjectConfig, kind string, allowCreate bool) (context.Context, config.ProjectConfig, error) {
	if o.D.State == nil {
		return ctx, cfg, Wrap("identity", cfg.Slug, errors.New("ownership store is required"), "Restore the StageServe identity ledger.")
	}
	if err := lifecycleContextErr(ctx, cfg); err != nil {
		return ctx, cfg, err
	}
	rec, recordErr := o.D.State.Load(cfg.Slug)
	if recordErr != nil && !errors.Is(recordErr, state.ErrNotFound) {
		return ctx, cfg, Wrap("identity", cfg.Slug, recordErr, "Restore or explicitly recover the retained identity ledger.")
	}
	if recordErr == nil && (rec.ProjectID == "" || rec.InstallationID == "") {
		return ctx, cfg, Wrap("identity", cfg.Slug, state.ErrLegacyState, "Explicitly export/import this project before runtime mutation.")
	}
	id, err := o.D.State.IdentityForPath(cfg.Dir)
	if errors.Is(err, state.ErrNotFound) && allowCreate && errors.Is(recordErr, state.ErrNotFound) {
		id, err = o.D.State.CreateIdentity(cfg.Dir, cfg.Slug)
	}
	if err != nil {
		return ctx, cfg, Wrap("identity", cfg.Slug, err, "Restore or explicitly recover the retained identity ledger.")
	}
	if id.Slug != cfg.Slug || (recordErr == nil && (rec.ProjectID != id.ProjectID || rec.InstallationID != id.InstallationID)) {
		return ctx, cfg, Wrap("identity", cfg.Slug, state.ErrOwnerMismatch, "Inspect immutable identity and canonical path before recovery.")
	}
	if !allowCreate && !id.Registered {
		return ctx, cfg, Wrap("identity", cfg.Slug, state.ErrOwnerMismatch, "Attach this retained project before changing its runtime.")
	}
	shared, err := o.D.State.InstallationIdentity(cfg.StackHome)
	if err != nil {
		return ctx, cfg, Wrap("installation-identity", cfg.Slug, err, "Restore the installation resource ledger.")
	}
	for _, identity := range []state.Identity{id, shared} {
		pending, e := o.D.State.PendingOperations(identity.ProjectID)
		if e != nil {
			return ctx, cfg, Wrap("journal", cfg.Slug, e, "Inspect pending operation recovery.")
		}
		if len(pending) > 0 {
			return ctx, cfg, Wrap("journal", cfg.Slug, state.ErrCollision, "Recover the pending operation before retrying.")
		}
	}
	cfg.ComposeProjectName = "stage-" + id.ProjectID
	cfg.DatabaseVolume = cfg.ComposeProjectName + "-db-data"
	cfg.SharedGateway.ComposeProjectName = "stage-shared-" + shared.InstallationID
	hash := ""
	op, err := o.D.State.BeginOperation(id.ProjectID, id.Revision, kind, hash)
	if err != nil {
		return ctx, cfg, Wrap("journal", cfg.Slug, err, "Recover pending operations before retrying.")
	}
	if recordErr == nil {
		op.PriorConfiguration, _ = json.Marshal(rec)
		if err = o.D.State.SaveOperation(op); err != nil {
			return ctx, cfg, Wrap("journal", cfg.Slug, err, "Inspect operation persistence before retrying.")
		}
	}
	tx := &lifecycleTransaction{store: o.D.State, identity: id, operation: op}
	if recordErr == nil {
		copy := rec
		tx.record = &copy
	}
	sharedOp, err := o.D.State.BeginOperation(shared.ProjectID, shared.Revision, "shared-"+kind, hash)
	if err != nil {
		_ = tx.phase("rolled_back")
		return ctx, cfg, Wrap("journal", cfg.Slug, err, "Recover the installation operation before retrying.")
	}
	sharedTx := &lifecycleTransaction{store: o.D.State, identity: shared, operation: sharedOp, installation: true}
	ctx = context.WithValue(ctx, transactionKey(false), tx)
	ctx = context.WithValue(ctx, transactionKey(true), sharedTx)
	return ctx, cfg, nil
}
func (o *Orchestrator) runOwned(ctx context.Context, cfg config.ProjectConfig, kind string, allowCreate bool, action func(context.Context, config.ProjectConfig) error) error {
	ctx, cfg, err := o.beginTransaction(ctx, cfg, kind, allowCreate)
	if err != nil {
		return err
	}
	tx, shared := transactionFrom(ctx), installationTransactionFrom(ctx)
	if err = tx.phase("apply"); err == nil {
		err = shared.phase("apply")
	}
	if err == nil {
		err = action(ctx, cfg)
	}
	if err == nil {
		if err = shared.commit(); err != nil {
			return Wrap("commit-installation", cfg.Slug, err, "Recover the durable installation commit before retrying.")
		}
		if err = tx.commit(); err != nil {
			return Wrap("commit-project", cfg.Slug, err, "Recover the durable project commit before retrying.")
		}
		return nil
	}
	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()
	var cleanupErrors []error
	if kind == "down" || kind == "detach" || kind == "restart" {
		cleanupErrors = append(cleanupErrors, errors.New("runtime action did not complete; inspect pending recovery"))
	}
	if tx.needsCleanup {
		if e := tx.phase("rollback"); e != nil {
			cleanupErrors = append(cleanupErrors, e)
		}
		if e := o.stopProject(cleanupCtx, cfg, false); e != nil {
			cleanupErrors = append(cleanupErrors, e)
		}
		if e := o.syncSharedGateway(cleanupCtx, cfg, ""); e != nil {
			cleanupErrors = append(cleanupErrors, e)
		}
		if e := removeEnvFile(cfg); e != nil {
			cleanupErrors = append(cleanupErrors, e)
		}
	}
	for _, t := range []*lifecycleTransaction{tx, shared} {
		for _, r := range t.operation.Resources {
			if r.Planned {
				cleanupErrors = append(cleanupErrors, errors.New("resource creation remains unresolved"))
			}
		}
	}
	// Installation resources remain retained even when the project fails.
	if len(cleanupErrors) == 0 {
		if e := shared.commit(); e != nil {
			cleanupErrors = append(cleanupErrors, e)
		}
	}
	if len(cleanupErrors) > 0 {
		for _, t := range []*lifecycleTransaction{tx, shared} {
			if t.finished {
				continue
			}
			if e := t.store.SaveIdentity(t.identity, t.identity.Revision); e != nil {
				cleanupErrors = append(cleanupErrors, e)
			}
			t.operation.Leftovers = nil
			seen := map[string]bool{}
			for _, r := range append(append([]state.ResourceRecord{}, t.operation.Resources...), t.identity.Resources...) {
				key := resourceKey(r.Kind, r.ID, r.CreationOperation)
				if (r.Planned || (!r.Deleted && r.Kind != "volume")) && !seen[key] {
					t.operation.Leftovers = append(t.operation.Leftovers, r)
					seen[key] = true
				}
			}
			// Journal resource/leftover collections are disjoint.
			if len(t.operation.Leftovers) > 0 {
				filtered := []state.ResourceRecord{}
				for _, r := range t.operation.Resources {
					if !r.Planned && (r.Deleted || r.Kind == "volume") {
						filtered = append(filtered, r)
					}
				}
				t.operation.Resources = filtered
			}
			if e := t.phase("degraded"); e != nil {
				cleanupErrors = append(cleanupErrors, e)
			}
		}
		return errors.Join(append([]error{err}, cleanupErrors...)...)
	}
	if e := tx.store.SaveIdentity(tx.identity, tx.identity.Revision); e != nil {
		return errors.Join(err, e)
	}
	if e := tx.phase("rolled_back"); e != nil {
		return errors.Join(err, e)
	}
	return err
}
func (o *Orchestrator) Up(ctx context.Context, cfg config.ProjectConfig) error {
	if err := o.validateRuntimeCapabilities(cfg); err != nil {
		return err
	}
	if err := ValidateRuntimeAssets(cfg); err != nil {
		return err
	}
	return o.runOwned(ctx, cfg, "up", true, o.up)
}
func (o *Orchestrator) Attach(ctx context.Context, cfg config.ProjectConfig) error {
	if err := o.validateRuntimeCapabilities(cfg); err != nil {
		return err
	}
	if err := ValidateRuntimeAssets(cfg); err != nil {
		return err
	}
	return o.runOwned(ctx, cfg, "attach", true, o.attach)
}
func (o *Orchestrator) Down(ctx context.Context, cfg config.ProjectConfig, removeVolumes bool) error {
	return o.runOwned(ctx, cfg, "down", false, func(ctx context.Context, cfg config.ProjectConfig) error { return o.down(ctx, cfg, removeVolumes) })
}
func (o *Orchestrator) Detach(ctx context.Context, cfg config.ProjectConfig) error {
	return o.runOwned(ctx, cfg, "detach", false, o.detach)
}
func (o *Orchestrator) RestartService(ctx context.Context, cfg config.ProjectConfig, service string) error {
	if strings.TrimSpace(service) == "" {
		return Wrap("restart-service", cfg.Slug, errors.New("service name is required"), "Choose a specific service before restarting it.")
	}
	return o.runOwned(ctx, cfg, "restart", false, func(ctx context.Context, cfg config.ProjectConfig) error { return o.restartService(ctx, cfg, service) })
}
func routesWithoutProject(routes []gateway.Route, slug string) []gateway.Route {
	out := []gateway.Route{}
	for _, r := range routes {
		if r.Slug != slug {
			out = append(out, r)
		}
	}
	return out
}
