# Implementation Plan: Apple-only guided StageServe

**Branch**: `codex/012-apple-only-experience` | **Date**: 2026-09-13 | **Spec**: [spec.md](spec.md)
**Input**: `specs/012-apple-only-experience/spec.md`
**Status**: Revised after eleven-finding audit, 2026-09-13; implementation and G0/GUX/G1–G4 remain open.

## Summary

Converge existing Apple-only runtime work and guided terminal work into one trustworthy local PHP/MariaDB product. Prove one runtime vertical slice, then ownership/recovery and multi-project safety, complete the TUI, and release with install/update/migration evidence. Preserve shared core and existing Go/Charm code. No v1 GUI, Docker fallback, new library or DNS extraction project.

[Roadmap](../../docs/roadmap.md) owns milestone estimates; [research](research.md) owns decisions and current evidence; [tasks](tasks.md) owns execution order. Old plans are historical, not parallel authorities.

## Technical Context

**Language/Version**: Go 1.26; go.mod toolchain 1.26.2. Review host runs Go 1.27.1; release checks must use declared supported toolchain.
**Primary Dependencies**: existing Cobra, Bubble Tea, Bubbles, Huh, Lip Gloss, x/sys; Apple `container` subprocess adapter. Remove unused Docker SDK/Compose code by release.
**Storage**: existing atomic JSON project records, generated env/routing files, named Apple database volumes, independent identity/retained-resource ledger and apply journal; no new database service for StageServe itself.
**Testing**: Go unit/contract/golden tests, fake subprocess tests derived from real versioned CLI fixtures, real Apple integration and PTY/manual acceptance, installer smoke scripts.
**Target Platform**: Apple silicon; macOS 26+ only where explicitly tested with a pinned Apple CLI release.
**Project Type**: local CLI/TUI application, no HTTP management API.
**Performance Goals**: cached dashboard interactions target <=200 ms; long operations asynchronous and cancellable with progress, health deadline default 120 s; measure runtime service/memory cost on the proposed 16 GiB/two-project minimum host; detailed CPU/RAM/disk/log/cancel bounds in release-qualification.md and ux-validation.md. These targets are unverified until M1/M3.
**Constraints**: one stack initially; offline operation uses cached images; no silent state adoption, no new dependencies, local endpoint exposure, existing project data preserved.
**Scale/Scope**: one operator/Mac; validate two concurrent projects and 20 start/stop cycles. Do not imply arbitrary-scale support.

## Constitution Check

Pre-research and post-design checks against version 3.1.0:

- [x] Ease of use: one guided entry, visible defaults, no implementation vocabulary required; alternate CLI retained.
- [x] Reliability: exact precedence and runtime selector, shared planner, explicit legacy schema handling and compatible updates.
- [x] Robustness: exact ownership, data-preserving stop, scoped locks, rollback/reconciliation, local routing and isolation gates.
- [x] Documentation parity: milestone tasks cover README, contracts, help, mockups, design and agent guidance.
- [x] Operational validation planned: startup/status/logs/teardown, failure/cancel and real browser/terminal gates; missing live runtime explicitly recorded.
- [x] Apple-only scope: unsupported runtime fails closed; no Docker compatibility promise; GUI deferred.

These checkboxes validate the design, not implementation readiness. G0–G4 remain open until evidence is collected. No constitutional exception is required.

## Project Structure

### Documentation (this feature)

`spec.md`, `plan.md`, `research.md`, `data-model.md`, `quickstart.md`, `contracts/application-contract.md`, `contracts/project-runtime-contract.md`, `contracts/release-qualification.md`, `ux-validation.md`, `tasks.md`, `checklists/requirements.md`, `legacy-lessons.md`, `analysis.md`.

### Source Code (repository root)

| Path | Ownership after implementation |
|---|---|
| `cmd/stage/commands/` | Cobra commands and adapters to application actions; no runtime subprocess logic |
| `core/guidance/` | context/planner, semantic view model, existing Bubble Tea shell; no direct infrastructure decisions |
| `core/config/`, `core/project/`, `core/state/` | settings, identity, persistence/migration |
| `core/lifecycle/`, `core/runtime/` | orchestration, operation boundaries and runtime contract |
| `infra/applecontainer/` | pinned Apple CLI invocation/inspection, mounts, owned resources and observed addresses |
| `infra/gateway/`, `stacks/20i/`, `docker/` | route generation, Apple manifests and reusable OCI build inputs; directory name alone is not a runtime dependency |
| `platform/dns/`, `platform/tls/`, `platform/ports/` | host integration, previewed host changes, port locking |
| `observability/status/`, `observability/logs/` | reconciled status and bounded streams |
| `docs/design/`, `mockups/` | design reference and fixtures, never lifecycle authority |
| `scripts/tests/`, `.github/workflows/`, `install.sh` | install/package/update checks and release matrix |

**Structure Decision**: evolve existing modules, extract only concrete duplicated action wiring if required. No speculative daemon, plugin system, remote server or new GUI repository.

## Phase 0: Research and convergence (M0)

Capture dirty baseline and reconcile tests without weakening product assertions. Use an Apple test host to capture exact flags/JSON/versions. Resolve networking proof, legacy-state distinction, actual readiness and DNS provider strategy. Gate G0 stops runtime delivery if these cannot be proven. Deliver recorded topology and compatibility contract before broad UI integration.

## Phase 1: Design and implementation boundaries (M1/M2)

Runtime observations feed lifecycle/state; project endpoints feed generated PHP/gateway configs. Lifecycle is the single owner of operation ordering/locks and persistence commit boundaries. Actions validate ownership immediately before mutation and reconcile after errors. Health uses persisted/loaded manifests and total deadline, independent of in-memory cache. Rollback gets a fresh bounded cleanup context and reports leftovers.

TUI adapters call the same domain operations as CLI. Refresh publishes a new snapshot only after reinspection; stale plans cannot authorize destructive actions. Logs use bounded buffers/cancellation rather than unbounded body strings. No OpenAPI document is needed: this is in-process/CLI functionality, specified in the application contract.

## Phase 2: Delivery sequencing (M3/M4)

[Tasks](tasks.md) map each story to code, checks and gates. M3 fixture preparation/study follows T004; full dashboard implementation waits for GUX and G0, and live acceptance waits for G2. Release requires clean install, update/rollback and migration rehearsal, current help/design/docs, and no unsupported Docker execution dependencies. No implementation is authorized merely by a checked planning checklist.

## Complexity Tracking

No constitution violations. Retained complexity: shared gateway to support stable multi-project local URLs, explicit Apple adapter for external CLI change, and separate semantic projections for interactive/non-interactive operation. Rejected: dual runtime support, simultaneous native GUI, generic Compose parser and standalone DNS product. Revisit topology at G0/G1 if a simpler Apple design proves equivalent isolation and routing.

## Revised decisions and research ownership

The [project/runtime contract](contracts/project-runtime-contract.md) is normative
for settings transactions, retained ledger, app .env fallback/probes, private-network
matrix and mixed TLS. The [release contract](contracts/release-qualification.md)
is normative for actual artifact qualification, post-write compatibility, offline
and host lifecycle. [UX validation](ux-validation.md) defines GUX and accessible
text; it is a protocol, not fabricated participant evidence.

T038 fixes authority/archive reproducibility now. T040 establishes durable identity
before G0; T041/T042 add settings/cached-resource proof to G1. T039 gates T024; T043
qualifies accessible text; T044 covers host lifecycle; T045 makes release gates
executable. Revised estimate is 9.5–17.5 engineering weeks, with G0/GUX/G1 re-estimates.

Post-design constitution check: all expanded obligations have explicit contract,
case and task ownership. Empirical feasibility remains a hard gate, not an exception.
No new runtime dependency is introduced by these planning repairs.
