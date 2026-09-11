package commands

import (
	"context"

	"github.com/peternicholls/stageserve/core/config"
	"github.com/peternicholls/stageserve/core/onboarding"
	"github.com/peternicholls/stageserve/infra/applecontainer"
)

func buildMachineReadinessResult(shared *SharedFlags, suffix string) (onboarding.CommandResult, error) {
	cfg, err := loadConfig(shared)
	if err != nil {
		return onboarding.CommandResult{}, err
	}
	checkSuffix := cfg.SiteSuffix
	if suffix != "" {
		checkSuffix = suffix
	}
	return onboarding.BuildResult(machineReadinessSteps(cfg, checkSuffix), nil, nil), nil
}

func buildMachineReadinessResultForConfig(cfg config.ProjectConfig) onboarding.CommandResult {
	return onboarding.BuildResult(machineReadinessSteps(cfg, cfg.SiteSuffix), nil, nil)
}

func machineReadinessSteps(cfg config.ProjectConfig, suffix string) []onboarding.StepResult {
	apple := applecontainer.Probe{}.Check(context.Background())
	return []onboarding.StepResult{
		applePlatformStep(apple),
		appleBinaryStep(apple),
		appleServiceStep(apple),
		onboarding.CheckStateDir(cfg.StateDir),
		onboarding.CheckRequiredFile("stack.shared_file", "Shared runtime file", cfg.SharedFile),
		onboarding.CheckRequiredFile("stack.project_file", "Project runtime file", cfg.StackFile),
		onboarding.CheckPort("port.80", 80),
		onboarding.CheckPort("port.443", 443),
		onboarding.CheckDNS(suffix),
		onboarding.CheckMkcert(),
	}
}

func applePlatformStep(readiness applecontainer.Readiness) onboarding.StepResult {
	if !readiness.ArchitectureOK || !readiness.OperatingSystemOK {
		return onboarding.StepResult{ID: "apple-container.platform", Label: "Apple Container platform", Status: onboarding.StatusError, Message: readiness.Message, Code: "unsupported-os"}
	}
	return onboarding.StepResult{ID: "apple-container.platform", Label: "Apple Container platform", Status: onboarding.StatusReady, Message: "Apple silicon and macOS 26 or later detected"}
}

func appleBinaryStep(readiness applecontainer.Readiness) onboarding.StepResult {
	if readiness.BinaryPath == "" {
		remediation := "Install Apple Container from its signed release package"
		return onboarding.StepResult{ID: "apple-container.binary", Label: "Apple Container", Status: onboarding.StatusNeedsAction, Message: "Apple Container is not installed", Remediation: &remediation}
	}
	return onboarding.StepResult{ID: "apple-container.binary", Label: "Apple Container", Status: onboarding.StatusReady, Message: "container found at " + readiness.BinaryPath}
}

func appleServiceStep(readiness applecontainer.Readiness) onboarding.StepResult {
	if !readiness.ServiceRunning {
		remediation := "Run container system start, then retry stage doctor"
		return onboarding.StepResult{ID: "apple-container.service", Label: "Apple Container service", Status: onboarding.StatusNeedsAction, Message: readiness.Message, Remediation: &remediation}
	}
	return onboarding.StepResult{ID: "apple-container.service", Label: "Apple Container service", Status: onboarding.StatusReady, Message: "Apple Container system service is running"}
}
