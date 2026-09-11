# Feature Specification: Apple-only guided StageServe

**Feature Branch**: `codex/012-apple-only-experience`  
**Spec Kit feature**: `012-apple-only-experience` (set `SPECIFY_FEATURE` for scripts)  
**Created**: 2026-09-11  
**Status**: Planned; implementation and live acceptance pending  
**Input**: Apple Containers only; consolidate GUI/TUI work, decide the product, review old specs for lessons, archive old planning, and plan the complete roadmap.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Run my first local site (Priority: P1)

A developer with an Apple silicon Mac opens a PHP project folder and uses `stage` to inspect prerequisites, preview settings and run a local site without reading a command sequence.

**Why this priority**: This is the smallest complete product value, dependent on a real working runtime.
**Independent Test**: On a clean supported host, from an unrelated project folder, create settings, run a PHP page that reads/writes MariaDB, open the displayed URL, stop and restart with the same data.
**Acceptance Scenarios**:
1. Given missing Apple software or stopped services, when opening StageServe, then report the precise blocker without starting Docker or modifying the host automatically.
2. Given no project settings, when setup is selected, then show resolved name, path, web folder and URL before writing; cancel leaves files unchanged.
3. Given a healthy stack, when starting, then report running only after required health and routed application checks pass.
4. Given an existing config, when starting from its folder, then preserve explicit settings and explain configuration origins.

### User Story 2 - Operate projects confidently (Priority: P1)

A returning developer sees project state, opens the site, follows logs and stops/restarts the selected project. CLI and TUI agree.

**Independent Test**: Run two differently named projects, switch selection, follow one service's logs, stop one project and verify the other site and both databases remain intact.
**Acceptance Scenarios**:
1. Given running state, when pressing Enter, then view logs without restarting or stopping the project.
2. Given concurrent projects, when operating on one, then scope is visible and unrelated routes/data are unchanged.
3. Given non-interactive or no-colour output, when invoking equivalent commands, then output remains useful and never prompts or emits TUI decoration into JSON.

### User Story 3 - Recover safely (Priority: P1)

A developer can understand partial startup, stale state, conflicts, unavailable dependencies and cancellation, then recover without losing data.

**Independent Test**: Inject one failure after each startup stage and a runtime restart; verify truthful status, resource ownership, useful recovery and no damage to another project.
**Acceptance Scenarios**:
1. Given failed startup, when it exits or is cancelled, then show retained resources and the failed stage; never leave a false healthy/attached record.
2. Given a stale action preview, when state changes before execution, then revalidate or refuse rather than execute against the old scope.
3. Given data removal, when no explicit scoped confirmation is supplied, then retain all data.
4. Given legacy Docker records, when read by the Apple product, then identify them as unsupported legacy state without adopting or deleting resources.

### User Story 4 - Install, update and retain my work (Priority: P2)

A developer installs a compatible binary/runtime-asset pair, understands supported host versions, and can update or return to a previous compatible release without losing project data.

**Independent Test**: From a clean user account install verified artifacts; run US1; upgrade with a project present; exercise interrupted download and rollback using a retained backup.
**Acceptance Scenarios**:
1. Given unsupported hardware, OS or runtime version, when checking readiness, then fail with actionable guidance before mutation.
2. Given update failure, when rolling back, then preserve config/data and restore a compatible executable/asset pair.
3. Given a Docker-era project, when following migration guidance, then source and settings are inventoried and data export/import is explicit; no Docker fallback is invoked by StageServe.

### Edge Cases

Paths with spaces/non-ASCII; duplicate slug or hostname; missing web folder; moved project; corrupt/newer schema; missing service; address change after restart; occupied ports; missing image/offline pull; DNS mismatch; certificate trust denial; permission failure; 52-column resize; terminal interruption; logs backpressure; repeated cancel; simultaneous CLI/TUI actions; unavailable debug service; partial update; multiple projects sharing routing.

## Operator Experience & Safety *(mandatory)*

### Friction & Entry Points

Primary entry is bare `stage` in an interactive project folder. Direct commands keep their documented intent (`setup`, `init`, `up`, `status`, `logs`, `down`, `doctor`). Preserve `--notui`, `--cli`, `STAGESERVE_NO_TUI=1`, `NO_COLOR` and structured output where supported. No new generic runtime picker, GUI setup path or arbitrary Compose importer. Project selection and restart UI use existing domain actions; any missing command exposure needs explicit contract tests.

### Configuration & Precedence

CLI flags -> project `.env.stageserve` -> shell environment -> stack `.env.stageserve` -> defaults. `STAGESERVE_RUNTIME` absent or `apple-container` selects Apple; other values fail before mutation. `.test` remains the default suffix. Existing explicit suffixes and paths remain unchanged until validated migration. Generated env files are outputs, never an undocumented higher-precedence configuration source.

### State, Isolation & Recovery

StageServe owns project IDs, selected scope, records, generated routing and explicitly named runtime resources. Runtime IDs/addresses are observations. Stop retains database volumes; deletion is separate. Atomic records, operation locking and scope revalidation protect concurrent commands. Unknown legacy runtime identity must remain legacy/unknown, never silently Apple. Recovery proceeds from inspection to explicitly scoped repair.

### Documentation Surfaces

Update README, runtime contract, architecture, installer/onboarding, migration, command help, design guides, mockups, scoped agent instructions and active Spec Kit artifacts together as milestones land. Historical sources retain their original completion statements but have no authority over current readiness.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Support only Apple `container` on tested Apple silicon/macOS combinations; reject unsupported runtime selection before side effects.
- **FR-002**: Preview effective project settings and file path before create/edit; preserve deterministic precedence and cancellation without writes.
- **FR-003**: Start required services in dependency order; prove routed PHP/database operation and bounded health waits before running/attached success.
- **FR-004**: Preserve project identity and database data across stop/start and application updates; separate stop, unregister and data deletion.
- **FR-005**: Show one selected project, verdict, evidence and safe next action through a shared planner and application services.
- **FR-006**: Provide working logs, status, browser-open and lifecycle actions without renderer-owned business logic; Enter on running opens logs.
- **FR-007**: Support two simultaneous projects with isolated state/data and scoped cleanup, including shared-routing failure recovery.
- **FR-008**: Expose truthful desired/observed state and actionable errors for partial startup, stale state and unavailable runtime; recheck after actions.
- **FR-009**: Require explicit scoped confirmation for data deletion and host DNS/trust changes; never reset unrelated resources.
- **FR-010**: Preserve direct CLI, plain text and JSON semantics; no prompts in non-TTY, no UI text in JSON, no colour-only information.
- **FR-011**: Meet 52/80-column, resize, keyboard, cancellation and no-colour acceptance scenarios without losing essential action/scope information.
- **FR-012**: Ship verified binary/asset compatibility, host/runtime version gates, update recovery and explicit legacy migration guidance.
- **FR-013**: Prove local DNS/routing and optional TLS independently; show actual scheme/port and never silently modify configured suffixes.
- **FR-014**: Keep secrets out of diagnostics/JSON/support output and default endpoints local; inventory every published port.
- **FR-015**: Preserve legacy planning with provenance and lessons; keep one active roadmap and trace all requirements to tasks and evidence.

### Key Entities

Project settings; project identity/record; observed service; operation and result; readiness report; next action plan; route; persistent data reference; compatible release/asset pair. See [data model](data-model.md).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Three representative first-use trials complete setup/start/visit/stop without undocumented command ordering; record time, blockers and corrections. Target <=10 minutes excluding prerequisite/image downloads, with no unresolved critical usability blocker.
- **SC-002**: Twenty repeated start/stop cycles retain a database sentinel, stable project identity and no leaked owned resources; every successful start passes routed application checks.
- **SC-003**: All failure/cancellation scenarios in the acceptance matrix produce truthful status and preserve unrelated project data/routes; zero unintended deletion.
- **SC-004**: All canonical planner situations pass TUI/text semantic parity and JSON schema tests; 52/80-column real-terminal review retains visible scope and exit/cancel controls.
- **SC-005**: Two projects survive independent operations and shared-route recovery; their unique HTTP and database sentinels never cross.
- **SC-006**: Clean install, update, interrupted update and rollback pass on the declared supported matrix before release; compatibility range and evidence are published.
- **SC-007**: Every archived file matches its recorded checksum; every FR maps to implementation and validation tasks; no old spec remains an active backlog.

## Assumptions

One experienced Go developer-equivalent with access to an Apple silicon test Mac; existing libraries are reused, no new dependencies authorized. PHP/MariaDB development is the initial use case. Network access is required for initial images/tool installation. TUI is the selected v1 interface; GUI evaluation is conditional post-v1. Runtime version selection and topology proof are explicit M0/M1 experiments, not implicit support promises. This planning task does not implement or release the product.
