<!--
Sync Impact Report
Version: 2.0.1 -> 3.0.0 (breaking runtime/platform contract)
Authority: user direction, 2026-09-11: Apple Containers only.
Preserved: ease of use, reliability, robustness, friction removal.
Changed: Apple-only runtime, shared TUI/CLI semantics, explicit legacy migration.
Synced: spec/plan templates, README authority notice, active spec 012,
        docs/product-direction.md, docs/roadmap.md, agent context.
Historical specs are archived; old implementation docs carry transition notices.
-->

# StageServe Constitution

## Core Principles

### I. Ease Of Use Is A Product Requirement
The primary operator experience MUST stay simple enough to use from memory.
Changes to commands, GUI wrappers, and future attach or detach commands MUST
preserve the shortest obvious path for the common workflow or provide a clearly
documented migration in the same change.
Any added friction MUST be justified by the pain it removes, and any deferred
complexity MUST be called out explicitly in the plan and docs.
Routine operations MUST prefer sensible defaults, low setup overhead, and
minimal manual coordination between shell, GUI, and automation entry points.

Rationale: this stack only creates value when local project startup and
inspection remain easier than hand-rolled local container workflows.

### II. Reliability Must Be Boring And Predictable
The same input MUST produce the same operational result across shell commands,
automation wrappers, and repeated runs. Configuration precedence MUST remain
deterministic, documented, and operator-visible. New variables MUST use a single
canonical name, declare a default or required state, and avoid hidden coupling
to undocumented environment state. User-visible behavior changes MUST update all
affected operator surfaces in the same delivery unit.

Rationale: the stack is infrastructure, not a novelty feature; confidence comes
from repeatable behavior and documentation that matches reality.

### III. Robustness Must Hold Under Real Failure
Automation MUST handle partial startup, stale state, shared-service drift,
missing dependencies, and project-level breakage without corrupting unrelated
projects. Isolation of code mounts, naming, persistent state, and database data
is mandatory unless a resource is explicitly defined as shared infrastructure.
Changes affecting startup, routing, DNS, ports, attach or detach flows, or
teardown MUST expose actionable diagnostics through status, logs, health checks,
or explicit recovery instructions. Ambiguous state MUST be reported, not hidden.

Rationale: local infrastructure fails in messy ways, so resilience and recovery
clarity matter more than optimistic happy-path behavior.

### IV. Remove Pinch Points And User Friction
Features and process changes MUST actively reduce recurring operator pain rather
than shift it elsewhere. Any new step, prompt, manual sync, or stateful
exception MUST be justified in the plan, along with the friction it removes and
the simpler alternative considered. Runtime assets and the binary MUST have an explicit compatibility relationship;
user project settings and persistent data MUST survive asset and binary updates.

Rationale: this project exists to remove the repeated annoyances that make local
stack management slower, harder to remember, or easier to get wrong.

## Operational Constraints

- The supported runtime is Apple `container` only, on Apple silicon and a tested
  macOS 26-or-later release. The supported CLI version range MUST be recorded
  from live validation. Docker Desktop, Compose, and Docker socket compatibility
  are not supported product paths or fallback mechanisms.
- The primary interface is a guided TUI backed by the same application services
  as direct CLI, plain text, and structured JSON output. Renderers MUST NOT own
  lifecycle truth. A separate GUI requires a later evidence-backed decision.
- A supported project can be operated from its own folder, independently of the
  source checkout. Configuration precedence is CLI flags -> project
  `.env.stageserve` -> shell environment -> stack `.env.stageserve` -> defaults.
- Persistent project identity and database data MUST NOT be reinterpreted as
  Apple resources merely because an old record lacks a runtime identifier.
  Legacy Docker data requires an explicit backed-up migration, not adoption.
- Development credentials and endpoint exposure MUST remain clearly labeled
  development-only. Default exposure MUST stay local; network isolation and
  routing must be demonstrated on the selected Apple runtime release.
- Shared infrastructure MUST define bootstrap, steady-state, teardown and recovery.
  Project stop/removal MUST preserve other projects and database volumes unless
  the operator explicitly requests scoped data deletion.
- OS-level DNS or trust changes MUST show exact effects and require operator
  authorization; normal project actions MUST NOT require repeated privilege prompts.

## Delivery Workflow & Quality Gates

- Every feature specification MUST identify the operator friction being removed
  or introduced, affected commands and interfaces, configuration precedence,
  state or isolation impact, and the documentation surfaces that need updating.
- Every implementation plan MUST pass a Constitution Check covering ease of use,
  reliability, robustness, friction removal, and operational validation.
- Every task list MUST include the work needed to keep docs and alternate entry
  points aligned when behavior changes, plus validation for any claimed
  reduction in operator friction.
- Changes to Apple runtime definitions, images, routing, or automation MUST be
  validated against startup, status/inspection, teardown, and at least one
  failure path relevant to the change. If validation cannot be run, the gap MUST
  be recorded explicitly.
- Complexity that violates this constitution MAY be approved only when the plan
  records the violation, the simpler rejected option, and the reason the extra
  complexity is necessary now.

## Governance

This constitution supersedes conflicting workflow guidance in repository docs and
Speckit templates. Amendments MUST update this file and any affected templates or
operator docs in the same change.

Versioning policy for this constitution follows semantic versioning:

- MAJOR: remove a principle, redefine a principle incompatibly, or weaken a
  governance requirement in a materially different way.
- MINOR: add a new principle or materially expand project-wide obligations.
- PATCH: clarify wording, tighten examples, or make non-semantic editorial fixes.

Compliance review expectations:

- Specs MUST show operator-facing impact, especially any friction removed,
  added, or deferred.
- Plans MUST document how the work satisfies the Constitution Check and why any
  added complexity is justified.
- Tasks and implementation reviews MUST confirm documentation parity, the
  required validation scope, and the operator path for recovery from failure.
- Unresolved non-compliance MUST be treated as a blocker until explicitly
  justified and accepted in the plan.

**Version**: 3.0.0 | **Ratified**: 2026-04-01 | **Last Amended**: 2026-09-11
