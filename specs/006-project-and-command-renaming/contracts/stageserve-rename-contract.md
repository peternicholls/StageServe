# Rename Contract: StageServe And `stage`

## Status

Proposed contract for Gate A review.

This file exists to freeze the rename boundaries before implementation. No public cutover work should proceed until this contract is accepted.

## External Identity Contract

- Product name: `StageServe`
- Canonical CLI command: `stage`
- Searchability note: active docs should use a concise transitional phrase such as `StageServe (formerly Stacklane)` only where discoverability materially helps operators

## Internal Naming End-State Contract

The final state for this spec is:

- no active `stacklane` references remain in code or maintained documentation
- `stacklane` survives only in clearly historical notes, archival material, or explicitly labeled migration history

This includes repository-owned internal surfaces such as:

- `STACKLANE_*`
- `.env.stacklane`
- `.stacklane-state`
- package, module, path, and layout names that still carry the old identity

Temporary staging rule:

- these surfaces may remain temporarily during early dry-run and cutover phases if needed to reduce risk
- they are not acceptable as the final closeout state for this spec

Runtime prefix rename:

- `stln-*` must be renamed to `stage-*` by spec closeout
- this includes Docker Compose project names, network names, and volume names (e.g. `stln-shared` → `stage-shared`, `stln-<slug>` → `stage-<slug>`)
- no live runtime migration is required; this is a local-development-only product and a clean `docker system prune` is acceptable before first use under the new prefix

## Output Contract

- The canonical command must preserve existing exit-code behavior for equivalent success, needs-action, and failure paths.
- Machine-readable and JSON output must remain clean on stdout.
- Any deprecation or compatibility messaging must go to stderr only.
- Help text and examples must present `stage` as canonical once cutover is complete.

## Compatibility Contract

Recommended default for Gate A discussion:

- Keep a temporary `stacklane` forwarding shim only if dry-run parity proves it behavior-safe.
- If parity is not clean, do not ship the shim.
- If the shim ships, it must be removed before this spec is closed so no active `stacklane` command path remains.
- If the shim ships, it must:
  - forward `stacklane <args>` to `stage <args>`
  - preserve exit codes
  - avoid stdout contamination for JSON or machine-readable modes
  - print concise deprecation guidance to stderr
  - have an explicit removal milestone before implementation begins

This keeps compatibility as a controlled transition decision rather than an assumed final-state feature.

## Distribution Contract

The following external surfaces are in scope for this phase:

- installed binary name
- installer output and retrieval targets
- release asset names and checksums
- shell completions
- active docs and normative specs
- CI smoke paths and release automation that invoke the canonical command
- active internal naming surfaces that still carry `stacklane`

The following surfaces require explicit classification before change:

- repo metadata and short description
- package-manager specific identities
- any external references that are not repository-owned but are linked from active docs

## Recommended Gate A Defaults

These defaults are the current recommended path unless dry runs uncover a blocker:

- use a hard switch by default: ship `stage` as canonical and do not keep a long-lived `stacklane` shim
- rename `.env.stacklane` to `.env.stageserve`
- rename `.stacklane-state` to `.stageserve-state`
- rename `STACKLANE_*` to `STAGESERVE_*`
- rename `stln-*` to `stage-*` (Docker Compose project names, network names, volume names)
- rename runtime-visible gateway headers and sentinels from `Stacklane`/`stacklane` to `StageServe`/`stageserve`
- rename the Go module path to `github.com/peternicholls/stageserve` only when repository identity is ready to support it

Rationale:

- these changes remove literal `stacklane` references from active code and docs
- they keep the highest-risk scope focused on names rather than runtime behavior changes
- they avoid widening the migration to every internal abbreviation unless there is independent value

## Non-Goals

- widen lifecycle behavior or change compose topology
- restore archived Bash wrappers as current behavior
- keep temporary compatibility surfaces longer than needed for the cutover rehearsals

## Dry-Run Gates Required By This Contract

Before public cutover, the following must be green:

1. Active-surface inventory and classification
2. Local binary/install rehearsal
3. Completion and shell-cache rehearsal
4. Docs/spec rehearsal
5. CI/release rehearsal
6. Rollback rehearsal
7. Final zero-active-reference rehearsal for active code and docs

## Approval Checklist

- [ ] External name and canonical command accepted
- [ ] Final no-active-`stacklane` end state accepted
- [ ] Shim policy accepted
- [ ] Distribution surfaces for this phase accepted
- [ ] Abort criteria and rollback owner named
- [ ] Dry-run evidence locations defined