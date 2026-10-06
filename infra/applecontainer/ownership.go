package applecontainer

import (
	"context"
	"encoding/json"
	"fmt"
	coreruntime "github.com/peternicholls/stageserve/core/runtime"
	"path/filepath"
	"strings"
)

func validateMutation(ctx context.Context, owner coreruntime.Ownership, record coreruntime.ResourceRecorder) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := owner.Validate(); err != nil {
		return err
	}
	if record == nil {
		return fmt.Errorf("runtime mutation requires durable resource recorder")
	}
	return nil
}
func ownershipLabels(o coreruntime.Ownership, role string) []string {
	return []string{"--label", "io.stageserve.installation-id=" + o.InstallationID, "--label", "io.stageserve.project-id=" + o.ProjectID, "--label", "io.stageserve.operation-id=" + o.OperationID, "--label", "io.stageserve.scope=" + o.Scope, "--label", "io.stageserve.role=" + role}
}
func newResource(o coreruntime.Ownership, kind, name, role string) coreruntime.Resource {
	return coreruntime.Resource{Kind: kind, ID: name, Name: name, InstallationID: o.InstallationID, ProjectID: o.ProjectID, Role: role, CreationOperation: o.OperationID}
}
func hasRecorded(o coreruntime.Ownership, kind, id string) bool {
	for _, r := range o.Resources {
		if r.Kind == kind && r.ID == id && !r.Deleted {
			return true
		}
	}
	return false
}
func ownedObserved(o coreruntime.Ownership, kind, id, name, role string, labels map[string]string) (coreruntime.Resource, error) {
	for _, r := range o.Resources {
		if r.Kind != kind || r.ID != id || r.Deleted {
			continue
		}
		if r.Name == name && r.Role == role && r.InstallationID == o.InstallationID && r.ProjectID == o.ProjectID && r.CreationOperation != "" && labels["io.stageserve.installation-id"] == o.InstallationID && labels["io.stageserve.project-id"] == o.ProjectID && labels["io.stageserve.scope"] == o.Scope && labels["io.stageserve.role"] == role && labels["io.stageserve.operation-id"] == r.CreationOperation {
			r.Planned = false
			return r, nil
		}
	}
	return coreruntime.Resource{}, fmt.Errorf("resource %s %s has no matching recorded incarnation and observed owner; preserve it and inspect the ledger", kind, id)
}
func (m *Manager) observedService(ctx context.Context, name string) (*coreruntime.Service, error) {
	services, err := m.ListServices(ctx, "")
	if err != nil {
		return nil, err
	}
	var found *coreruntime.Service
	for _, s := range services {
		if s.ID == name || s.Name == name {
			if found != nil {
				return nil, fmt.Errorf("ambiguous service %s", name)
			}
			copy := s
			found = &copy
		}
	}
	return found, nil
}

// Apple 1.4.1 VolumeResource encodes id and configuration{name,labels}.
// Source: apple/container Sources/ContainerResource/Volume/VolumeResource.swift.
func (m *Manager) observedVolume(ctx context.Context, name string) (*coreruntime.Service, error) {
	out, err := m.Runner.Run(ctx, "volume", "list", "--format", "json")
	if err != nil {
		return nil, err
	}
	var items []struct {
		ID            string `json:"id"`
		Configuration struct {
			Name   string            `json:"name"`
			Labels map[string]string `json:"labels"`
		} `json:"configuration"`
	}
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, fmt.Errorf("parse volume list JSON: %w", err)
	}
	var found *coreruntime.Service
	for _, v := range items {
		if v.ID == "" || v.Configuration.Name != v.ID {
			return nil, fmt.Errorf("missing or inconsistent volume identity")
		}
		if v.ID == name {
			if found != nil {
				return nil, fmt.Errorf("ambiguous volume %s", name)
			}
			found = &coreruntime.Service{ID: v.ID, Name: v.Configuration.Name, Labels: v.Configuration.Labels}
		}
	}
	return found, nil
}
func (m *Manager) ensureOwnedVolume(ctx context.Context, name string, o coreruntime.Ownership, record coreruntime.ResourceRecorder) error {
	volume, err := m.observedVolume(ctx, name)
	if err != nil {
		return err
	}
	if volume != nil {
		resource, err := ownedObserved(o, "volume", volume.ID, name, "volume", volume.Labels)
		if err != nil {
			return err
		}
		return record(ctx, resource)
	}

	if hasRecorded(o, "volume", name) {
		return fmt.Errorf("recorded volume %s is missing; reconcile before creation", name)
	}
	resource := newResource(o, "volume", name, "volume")
	resource.Planned = true
	if err := record(ctx, resource); err != nil {
		return err
	}
	args := []string{"volume", "create", "--label", "io.stageserve.managed=true"}
	args = append(args, ownershipLabels(o, "volume")...)
	args = append(args, name)
	if _, err := m.mutateRun(ctx, args...); err != nil {
		return err
	}
	volume, err = m.observedVolume(ctx, name)
	if err != nil {
		return err
	}
	if volume == nil {
		return fmt.Errorf("created volume %s was not observed", name)
	}
	expected := o
	expected.Resources = []coreruntime.Resource{resource}
	observed, err := ownedObserved(expected, "volume", volume.ID, volume.Name, "volume", volume.Labels)
	if err != nil {
		return err
	}
	return record(ctx, observed)
}

func (m *Manager) mutateRun(ctx context.Context, args ...string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return m.Runner.Run(ctx, args...)
}

// Named mounts must be declared so the CLI cannot implicitly create unlabelled data.
func validateVolumeMounts(definition manifest, env map[string]string) error {
	declared := map[string]bool{}
	for _, name := range definition.Volumes {
		declared[expand(name, env)] = true
	}
	for _, service := range definition.Services {
		for _, mount := range service.Mounts {
			source, _, ok := strings.Cut(expand(mount, env), ":")
			if !ok || source == "" {
				return fmt.Errorf("service %s has an anonymous or invalid mount", service.Name)
			}
			if !filepath.IsAbs(source) && !strings.HasPrefix(source, ".") && !declared[source] {
				return fmt.Errorf("service %s named volume %s must be declared and owned", service.Name, source)
			}
		}
	}
	return nil
}
