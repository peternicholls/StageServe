# Implementation evidence

This ledger records implementation, separately from planning approval. Tasks and
live gates are completed only to the extent shown below. Full goal remains active.

## 2026-10-05 — T001 baseline

- Checkout: `codex/012-apple-only-experience`, HEAD `3998a3b`; existing planning,
  archive and instruction changes preserved. No runtime source diff at entry.
- Spec Kit prerequisites passed. Requirements checklist: 12 completed, zero open.
- Host: macOS 27.0.1 (26A434), arm64, Go 1.27.1.
- `go test -short ./...`: PASS, including the formerly failing lifecycle route
  identity and doctor Apple-readiness tests. Assertions were not weakened.
- Apple `container` absent from PATH and standard bin locations. Staticcheck absent.
- Official 1.4.1 signed installer downloaded for qualification, SHA-256
  `c0d2716afefbb194c93fae662e9cae7cc186bcbcf746816608ec673dd648a6a4` matches release
  metadata. `pkgutil --check-signature` confirms Apple Inc. Containerization
  developer signature and trusted notarization. No installation performed yet.
- System installation requires administrator authentication (`sudo -n` unavailable).
  CLI extraction/read-only inspection is being used for real help/schema discovery;
  it cannot establish running-runtime, networking or lifecycle acceptance.

## Open gates

G0 needs T002/T003 live CLI/topology evidence plus completed foundations. GUX needs
actual participants; no simulated results. G1–G4 remain open. No DNS/trust change,
project startup, retained-data mutation or release publication has been performed.

## 2026-10-06 — Foundation implementation

- Preserved the existing planning/archive, config and identity-ledger work.
- T005 completed: `core/state/migration_test.go` proves incompatible project
  records cannot be overwritten or removed, and verifies schema/owner refusal
  for installation records, identities and journals in a fresh store instance.
  `docs/migration.md` records non-adoption, backups and explicit DB export/import.
- T004 advanced: `init_contract_test.go` covers scripted JSON create/preserve/
  force/dry-run consent and no-write validation failures; guidance tests refuse
  TUI on redirected streams; config tests verify application DB origins and
  source preservation. T004 remains open for the complete versioned JSON,
  cancellation and destructive transition freeze, including --confirm-project.
- T002 preparation: added `scripts/capture-apple-container.py` and tests for
  bounded read-only capture, private files, refusal to overwrite evidence and
  failure reporting. This is tooling, not real CLI fixtures or live qualification.
- Rechecked CLI locations and `sudo -n`: no installed Apple CLI; administrator
  authentication is required. No runtime installation or host mutation occurred.
- PASS: `go test -short ./...`; focused race tests for state/config/guidance/
  commands; `go vet ./...`; `go build ./...`; `git diff --check`; Python capture
  tests (2); `scripts/verify-planning.py` (45 tasks, 23 requirements, 11 outcomes,
  88 hashes, 20 linked documents).
- Configured staticcheck lint unavailable: executable not installed. No dependency
  was added. G0, GUX and G1–G4 remain open; later runtime/UI/release work is gated
  by actual CLI topology and participant evidence.

### Continued foundation work

- T002 advanced with actual 1.4.1 version/help/unregistered-status fixtures in
  `infra/applecontainer/testdata/1.4.1/`. Downloaded the signed package from the
  official release, matched the recorded SHA-256, verified signature/notarization,
  and extracted it without installation. Help/version commands exited 0;
  unregistered system status exited 1; inventory requires a running service.
  No running/inspect/run/network-isolation acceptance is claimed.
- The readiness adapter now requires successful command execution and decoded
  `status: running`. Malformed, absent, null and unknown status fail closed.
  Real unregistered output plus clearly synthetic parser cases cover both paths.
- T004 advanced with entrypoint cancellation code 130, including signal-canceled
  subprocess fallback; explicit operational codes remain unchanged. Contract and
  entrypoint tests freeze that convention. Deadlines remain operational failures.
- T040 state foundations now serialize independent Store instances/processes
  using a retained bounded filesystem lock. Six real concurrent processes prove
  one installation/project identity; competing revisions and journal creation
  cannot both win. Journal phases, base revisions, resource provenance, retained
  tombstones and prior configuration are validated. Runtime label/ledger and
  lifecycle journal integration remain open, so T040 is not checked complete.
- Read-only review identified terminal journals hiding unresolved leftovers;
  terminal state must reject undeleted leftovers, including on persisted reload.
  Fixed and verified both save and reload paths; pending recovery continues to
  block new operations until leftovers are resolved or tombstoned.
- PASS after this increment: full short suite, focused race tests for state/
  Apple adapter/entrypoint, full vet/build, Python capture tests, planning checks
  and diff checks. After the review fix, repeated state race/vet and planning/
  diff checks passed. No new dependencies or commits; original tracked `stage`
  binary restored after a validation build overwrote it.

### Record ownership and JSON foundation

- T040 advanced: identity-bearing project records now require both UUIDs and an
  existing matching registered ledger, installation, canonical project path and
  slug. Missing ledger cannot be interpreted as an absent project. Existing
  UUIDs cannot be swapped/cleared and UUID-free records cannot be automatically
  adopted. Removal validates ownership and preserves the retained ledger.
  Fresh-store tampering, missing-ledger and unregister cases have regressions.
- Guidance context now reads through the validated StateStore rather than
  decoding raw records. Legacy/incompatible records produce warnings instead
  of being adopted as running; missing-state collection creates no directory.
- T004 JSON foundations: onboarding and hidden guidance inspection output carry
  schema version 1 while preserving existing field meanings/casing. Guidance
  scope includes the resolved directory/slug and only a validated recorded UUID.
  Config failures do not invent a scope. Supported surfaces and remaining gaps
  are recorded in contracts/json-output-contract.md. Status/logs JSON support
  and the complete action/destructive-transition freeze remain open.
- Read-only review found no additional defects in record/guidance validation.
  Full T040 still requires lifecycle journal integration, exact observed labels
  and retained-volume reconciliation; no checkbox or live gate is completed by
  these foundations alone.
- PASS after final changes: `go test -short ./...`; focused race suites for state,
  guidance, onboarding and commands; `go vet ./...`; `go build ./...`; planning
  verifier (now 21 linked documents); `git diff --check`. Staticcheck still absent.

### Durable commit/recovery and completed onboarding scope

- T040: added `LifecycleStore.CommitOperation` and `RecoverOperation` in
  `core/state/transaction.go`. Exact prior/target identity and registry projection
  are persisted in a committing journal before either projection is updated.
  Fresh-process recovery replays only matching prior/target bytes and refuses
  conflicts. Unregister retains owned data and tombstones. Fault-injection cases
  cover journal, identity, projection and terminal-write boundaries, including
  repeated recovery. Earlier unresolved runtime phases require manual recovery;
  they are not silently treated as completed storage commits.
- Review found an omitted-created-resource path; fixed commit/recovery validation
  to require every journal resource/leftover in the retained target ledger with
  exact provenance/deletion state. Committing payloads are immutable and conflicting
  identity writes fail. Installation corruption is now detected even with no
  identity records; orphaned journals cannot silently become a fresh installation.
- T004: unified `project_scope` for setup/doctor/init, using resolved dir/slug and
  only validated registered UUIDs. Config failures omit unresolved scope; invalid
  ledger/journal emits an explicit error and clears UUID. Init scope validation
  precedes settings writes. Symlink aliases compare canonical paths. Tests cover
  fresh project, dry-run/write, config failure, registered/identity-only state,
  missing/corrupt ledger/installation/journal, aliases and credential exclusion.
- T002 parser comparison: pinned 1.4.1 source uses nested `configuration.labels`
  and `status.state`. Parser now preserves labels, matches exact project labels
  instead of name prefixes and refuses missing/conflicting IDs. Tests are explicitly
  synthetic source-based coverage, not live inspect/run fixtures; arbitrary network
  interfaces are not selected as endpoints. UUID ownership/runtime mutation callbacks
  and installation-scoped shared-resource ledger remain to be integrated in lifecycle.
- PASS: full short suite; focused race suites for state/Apple adapter/onboarding/
  commands; full vet/build; planning verifier and diff check. T004/T040 and all
  live gates remain open for their outstanding complete obligations.
