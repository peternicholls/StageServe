package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/onboarding"
	"github.com/peternicholls/stageserve/core/state"
)

// onboardingProjectScope reads ownership without creating an identity. A missing
// ledger is normal for a fresh project; damaged or inconsistent state is not.
func onboardingProjectScope(cfg config.ProjectConfig) (scope *onboarding.ProjectScope, err error) {
	scope = &onboarding.ProjectScope{Dir: cfg.Dir, Slug: cfg.Slug}
	defer func() {
		if err != nil {
			scope.ProjectID = ""
		}
	}()
	if _, err := os.Stat(cfg.StateDir); err != nil {
		if os.IsNotExist(err) {
			return scope, nil
		}
		return scope, err
	}
	store, err := state.NewStore(cfg.StateDir)
	if err != nil {
		return scope, err
	}
	record, err := store.Load(cfg.Slug)

	if err != nil && !errors.Is(err, state.ErrNotFound) {
		return scope, err
	}
	if err == nil && record.ProjectID != "" {
		recordedPath, pathErr := filepath.Abs(record.Project.Dir)
		if pathErr != nil {
			return scope, pathErr
		}
		recordedPath, pathErr = filepath.EvalSymlinks(recordedPath)
		if pathErr != nil {
			return scope, pathErr
		}
		resolvedPath, pathErr := filepath.Abs(cfg.Dir)
		if pathErr != nil {
			return scope, pathErr
		}
		resolvedPath, pathErr = filepath.EvalSymlinks(resolvedPath)
		if pathErr != nil {
			return scope, pathErr
		}
		if recordedPath != resolvedPath {
			return scope, fmt.Errorf("recorded project path does not match resolved configuration")
		}
		scope.ProjectID = record.ProjectID
	}
	identity, err := store.IdentityForPath(cfg.Dir)
	if err != nil && !errors.Is(err, state.ErrNotFound) {
		return scope, err
	}
	if err == nil && identity.Registered {
		if identity.Slug != cfg.Slug || (scope.ProjectID != "" && scope.ProjectID != identity.ProjectID) {
			return scope, state.ErrOwnerMismatch
		}
		scope.ProjectID = identity.ProjectID
	}
	if scope.ProjectID != "" {
		if _, err := store.PendingOperations(scope.ProjectID); err != nil {
			return scope, err
		}
	}
	return scope, nil
}

func onboardingScopeErrorStep() onboarding.StepResult {
	remediation := "Inspect and repair the recorded StageServe project state before retrying"
	return onboarding.StepResult{ID: "project.scope", Label: "Project identity", Status: onboarding.StatusError, Code: "project-state-invalid", Message: "Project identity could not be validated", Remediation: &remediation}
}

func onboardingConfigErrorResult() onboarding.CommandResult {
	return onboarding.BuildResult([]onboarding.StepResult{{ID: "project.config", Label: "Project settings", Status: onboarding.StatusError, Code: "project-config-invalid", Message: "Project settings could not be resolved"}}, nil, nil)
}
