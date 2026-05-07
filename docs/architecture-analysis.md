 StageServe Architecture Analysis
    
    What's Good (Keep These Patterns)
    
    1. Clean layered architecture: core/ (domain logic) -> infra/ (external deps) -> cmd/ (CLI layer). Interfaces are defined in the infra packages, implementations depend on core.
    
    2. Interface-based abstractions: DockerClient, Composer, GatewayManager, PortAllocator, StateStore — all well-defined and testable via internal/mocks/.
    
    3. Atomic file operations: Both state/store.go and gateway/manager.go use temp-file + rename patterns for durability.
    
    4. Typed errors with context: lifecycle.StepError carries step name, project slug, and remediation — excellent DX.
    
    5. Config loader precedence chain: CLI > project .env.stageserve > shell env > stack .env.stageserve > defaults. Clean separation.
    
    Opportunities to Streamline & Improve
    
    1. Massive Duplication in Onboarding Commands (Medium Priority)
    
    setup.go, doctor.go, and init.go share nearly identical patterns:
    - Same 7-line resolveOutputMode() call + 12-line projection switch block repeated 3x
    - Same resolveOnboardingStateDir() boilerplate
    - Same setupExitError / doctorExitError / initExitError — structurally identical but declared separately
    
    Fix: Extract a shared onboardingCommand helper:
    - Single renderResult(mode, title, result) function that handles the projection switch
    - Single exit error type that takes the command name as a parameter (or just use a generic exitCode(int) error since the Error message doesn't need to be command-specific)
    
    2. Orchestrator is Both Flow Control AND Implementation (High Priority)
    
    orchestrator.go (511 lines) does too much. It has:
    - Flow orchestration (the 11-step Up/Down flows)
    - Helper functions for env file generation (writeEnvFile — 38 lines of env var construction)
    - Port resolution logic (resolveSharedGatewayPorts, sharedGatewayPortInUse, firstAvailableSharedGatewayPort)
    - Container lookup logic (findServiceContainer, observedRuntime)
    
    These helper functions should be in:
    - writeEnvFile -> core/config/envwriter.go (it's config materialization)
    - Port resolution -> platform/ports/ or a new core/lifecycle/ports.go
    - Container lookup -> infra/docker/ (it's docker-specific query logic)
    
    Fix: Split ~250 lines of helper functions out of orchestrator into their owning packages.
    
    3. SharedFlags is a Dumping Ground for ALL CLI Flags (High Priority)
    
    SharedFlags struct (18 fields) contains flags that are only relevant to specific commands:
    - DryRun, WaitTimeoutSecs, Profile, All — relevant to up/down
    - SiteName, DocRoot, PHPVersion, MySQL* — relevant to up/attach
    - StackHome — only used by doctor/setup
    
    This means every subcommand inherits and documents ~18 flags even though most aren't relevant.
    
    Fix: Move command-specific flags to their own structs with PersistentFlags() at the right parent level, or just keep them as local Flags() on each command.
    
    4. State Store Locking Granularity (Low Priority)
    
    Every Store method acquires the same sync.Mutex (s.mu), including Registry() which is read-only. Could use sync.RWMutex for read-heavy workloads (status reads, registry scans).
    
    Impact: Minor in practice since this is a CLI, not a long-running server, but worth noting.
    
    5. DNS Platform Module Incompleteness (Medium Priority)
    
    platform/dns/ only has:
    - macos.go — macOS implementation
    - linux.go — Linux implementation
    - common.go — shared helpers
    - types.go — interfaces
    
    But no NewProvider() constructor that picks the right implementation. The dns.NewProvider() call in dnssetup.go doesn't exist in the visible files. This means either the constructor is missing, or it's in a build-tagged file we haven't seen.
    
    6. Exit Code Pattern Needs Consolidation (Medium Priority)
    
    Three nearly-identical exit error types:
    - initExitError
    - setupExitError  
    - doctorExitError
    
    All implement ExitCoder with the same Error(), ExitCode(), and Silent() methods.
    
    Fix: A single generic type:
    go
    type exitCodeError struct{ code int }
    func (e *exitCodeError) Error() string     { return fmt.Sprintf("finished with exit code %d", e.code) }
    func (e *exitCodeError) ExitCode() int      { return e.code }
    func (e *exitCodeError) Silent() bool       { return true }
    func exitCodeWithCode(code int) error       { return &exitCodeError{code} }
    
    
    7. No go.sum visible / go 1.26 (Observation)
    
    The go.mod declares go 1.26 with toolchain go1.26.2 — that's future-dated (May 2026). Check if this intentional or if the system clock is off.
    
    8. compose.go has a Redundant Flag (Bug)
    
    Line 54-57 in compose.go:
    go
    if opts.Detach {
        args = append(args, "-d")
    } else {
        args = append(args, "-d") // we always want detach for orchestration
    }
    
    Both branches do the same thing. The Detach field is dead code. Either remove the field or wire it properly.
    
    9. Project Env Duplication (Medium Priority)
    
    Two separate env file rendering functions:
    - onboarding.WriteProjectEnv() — minimal file via stage init
    - commands.renderProjectEnvFile() — full file via stage up / stage attach
    
    They do different things but the concepts overlap (both write .env.stageserve). Merge or at least clarify why two code paths exist.
    
    10. Missing Integration Tests
    
    Tests are all unit-level. No e2e integration tests exist despite the e2e state file that's committed. The scripts/tests/ directory has installer smoke tests but no Go-based integration tests.
    
    Summary Priority Order
    
    | Priority | Issue | Impact |
    |----------|-------|--------|
    | High | Orchestrator bloat (511 lines) | Maintainability |
    | High | SharedFlags dumping ground | UX / CLI confusion |
    | High | Compose.go redundant flag | Actual bug |
    | Medium | Onboarding command duplication | 3x copy-paste |
    | Medium | Exit code types duplication | DRY violation |
    | Medium | Env file double rendering | Logic fragmentation |
    | Low | State store RWMutex | Minor performance |
    
    
    
    The architecture is fundamentally sound and well-thought-out. The main issues are duplication that crept in during active development and some boundary violations where helper functions stayed in the orchestrator instead of being placed in their owning packages.
    
    Want me to dive deeper into any of these, or start applying fixes?