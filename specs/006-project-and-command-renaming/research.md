# Research And Inventory: StageServe Rename

## Purpose

This file supports Gate A and Gate B of the rename workplan:

- Gate A: freeze the rename contract and non-goals
- Gate B: inventory active rename surfaces before editing

The inventory below is intentionally limited to active repository-owned surfaces. `previous-version-archive/` is historical reference only and should not drive the active rename plan.

## Planning Recommendation Confirmed

- Recommended planning style: gated hybrid
- Why: this rename spans command help, install, docs, CI, release, and rollback; a pure backlog is too easy to execute out of order and too weak at separating reversible dry runs from irreversible publication steps

## Active Surface Inventory

### Category 1: Rename Now

These are active user-facing surfaces where `stage` is currently canonical and will need explicit cutover work.

#### README And Operator Docs

- `README.md`
- `docs/installer-onboarding.md`
- `docs/runtime-contract.md`
- `docs/migration.md`
- `docs/contributing.md`
- `docs/architecture.md`

Observed patterns:

- StageServe is the current primary product name in headings and opening copy.
- `stage` is the current canonical command in install, setup, runtime, and migration examples.
- Installer and manual download examples currently fetch `stage_*` assets and place `stage` on `PATH`.

#### CLI Help And User-Facing Command Strings

- `cmd/stacklane/commands/root.go`
- `cmd/stacklane/commands/version.go`
- `cmd/stacklane/commands/init.go`
- command files with user-facing help comments or descriptions under `cmd/stacklane/commands/`

Observed patterns:

- root command uses `Use: "stage"`
- short and long descriptions are StageServe-branded
- some generated or inline next-step guidance still tells users to run `stage up`

#### Installer And Distribution Surfaces

- `install.sh`
- `README.md` install section
- `docs/installer-onboarding.md`
- release metadata and checksums still to be explicitly inventoried during CI/release rehearsal

Observed patterns:

- active docs describe `stage` binary download and install flow
- active docs refer to `stage-bin` and `stage` as build/install outputs

#### Active Spec And Planning Surfaces

- active specs that contain normative command literals under `specs/`
- especially the already-maintained workflow and onboarding artifacts that teach the operator contract

Observed patterns:

- spec 005 quickstart and task artifacts still teach `stage setup`, `stage doctor`, and `stage init`
- spec 004 research and contract artifacts still describe `stage` as the active command surface

### Category 2: Rename Later In This Spec

These surfaces may stay temporarily during early dry-run phases, but they must be renamed or removed before this spec is closed.

- `.env.stageserve`
- `.stageserve-state`
- `STAGESERVE_*`
- `stage-*` → `stage-*` (Docker Compose project names, network names, volume names in `core/config/loader.go` and compose YAML files)
- internal file-system and runtime names that derive from the old boundary above
- package/import paths such as `github.com/peternicholls/stage/...` if they remain repository-owned and active by closeout

Evidence observed in active code:

- command code and tests refer to `.env.stageserve` and `.stageserve-state`
- active runtime docs explicitly lock these surfaces as part of the current contract today, which means this spec must now replace that contract rather than preserve it

### Category 3: Mark Legacy Or Historical Only

These surfaces may still mention StageServe as part of historical analysis or migration framing, but they must not remain the canonical operator story after cutover.

- `docs/cli-naming-analysis.md`
- legacy migration framing that compares old and new names
- historical sections inside older specs where the reference is explanatory rather than normative

Requirement for this category:

- keep historical context where it is useful, but clearly label it as analysis, migration guidance, or legacy note

### Category 4: Archive Only

- `previous-version-archive/`

Requirement for this category:

- do not treat archive hits as active rename scope
- do not restore archived wrappers or GUI material as current behavior

## Command Namespace Risk Note

`stage` is a short, generic command name and must be treated as a first-class risk.

Current evidence from the repository search:

- no active runtime or doc surface currently uses `stage` as the canonical command
- the rename plan therefore has no in-repo ambiguity, but it does increase machine-level collision risk because the new command is much more generic than `stage`

Implications for dry runs:

- PATH shadowing must be tested explicitly
- stale shell hash and completion caches must be tested explicitly
- clean-machine and dirty-machine validation must stay separate

## Immediate Gaps Exposed By Inventory

1. The active spec folder for 006 originally lacked separate contract and research artifacts, which made it too easy to jump from idea to implementation without freeze points.
2. User-facing command references exist across README, installer docs, runtime docs, migration docs, and multiple maintained spec artifacts, so a narrow code-only rename would leave the operator contract inconsistent.
3. Internal `stage` surfaces are deeply present in both docs and tests, so the plan needs an explicit later-phase rename sweep rather than pretending the external cutover alone finishes the job.
4. The generic command name `stage` raises a machine-level collision risk that the previous `stage` command did not.

## Recommended Next Reads Before Any Rename Implementation

Use these as the first local anchors when implementation begins:

1. `cmd/stacklane/commands/root.go`
2. `install.sh`
3. `README.md`
4. `docs/installer-onboarding.md`
5. `docs/runtime-contract.md`

These are the smallest surfaces that directly control the canonical command story for operators.

## Gate A And Gate B Exit Criteria

### Gate A Exit Criteria

- rename contract accepted
- final no-active-`stage` end state accepted
- shim policy accepted or explicitly excluded
- rollback owner and abort criteria named

### Gate B Exit Criteria

- every active hit is classified as rename now, rename later in this spec, mark legacy, or archive only
- implementation starts from the smallest direct control surfaces rather than broad search-and-replace work
- namespace risk note for `stage` is part of the dry-run plan