package applecontainer

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Readiness is a read-only snapshot of Apple Container prerequisites.
type Readiness struct {
	ArchitectureOK    bool
	OperatingSystemOK bool
	BinaryPath        string
	ServiceRunning    bool
	Message           string
}

// Probe supplies injectable host discovery for readiness tests.
type Probe struct {
	Runner    Runner
	GOOS      string
	GOARCH    string
	OSVersion func(context.Context) (string, error)
	LookPath  func(string) (string, error)
}

// Check reads host and Apple Container state without starting services.
func (p Probe) Check(ctx context.Context) Readiness {
	result := Readiness{}
	goos := p.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	arch := p.GOARCH
	if arch == "" {
		arch = runtime.GOARCH
	}
	result.ArchitectureOK = arch == "arm64"
	if goos != "darwin" {
		result.Message = "Apple Container requires macOS 26 or later on Apple silicon"
		return result
	}
	versionFn := p.OSVersion
	if versionFn == nil {
		versionFn = commandOSVersion
	}
	version, err := versionFn(ctx)
	if err == nil {
		result.OperatingSystemOK = majorVersion(version) >= 26
	}
	lookup := p.LookPath
	if lookup == nil {
		lookup = exec.LookPath
	}
	result.BinaryPath, _ = lookup("container")
	if !result.ArchitectureOK || !result.OperatingSystemOK {
		result.Message = "Apple Container requires macOS 26 or later on Apple silicon"
		return result
	}
	if result.BinaryPath == "" {
		result.Message = "Install Apple Container, then run container system start"
		return result
	}
	runner := p.Runner
	if runner == nil {
		runner = CommandRunner{Bin: result.BinaryPath}
	}
	if err := checkServiceStatus(ctx, runner); err != nil {
		result.Message = "Apple Container is installed but its system service is not running"
		return result
	}
	result.ServiceRunning = true
	result.Message = fmt.Sprintf("Apple Container is ready at %s", result.BinaryPath)
	return result
}

// A successful subprocess is insufficient evidence of a ready service. Decode
// only the status discriminator; other versioned host/path fields are additive.
func checkServiceStatus(ctx context.Context, runner Runner) error {
	out, err := runner.Run(ctx, "system", "status", "--format", "json")
	if err != nil {
		return err
	}
	var status struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(out, &status); err != nil {
		return fmt.Errorf("invalid Apple Container system status: %w", err)
	}
	if status.Status != "running" {
		return fmt.Errorf("Apple Container system service is not running")
	}
	return nil
}

func commandOSVersion(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "sw_vers", "-productVersion").Output()
	return strings.TrimSpace(string(out)), err
}

func majorVersion(value string) int {
	major, _ := strconv.Atoi(strings.SplitN(strings.TrimSpace(value), ".", 2)[0])
	return major
}
