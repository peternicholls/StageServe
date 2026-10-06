# Tasks: Apple-only guided StageServe

**Input**: spec.md, plan.md, research.md, data-model.md, contracts/application-contract.md.
**Status**: Revised 2026-09-13. Runtime implementation and qualification tasks remain open; T038 records completed planning-authority/archive repair.
**Format**: `[ ] Tnnn [P] [USn] description (FR coverage; dependencies)`. `[P]` means eligible for separate ownership after dependencies, not permission to race on shared files.

## Phase 1 — Convergence and shared foundation (M0/G0)

- [x] T001 Capture current diff and recheck the two originally failing expectations (both pass after concurrent work) against the selected contract in `core/lifecycle/orchestrator_test.go` and `cmd/stage/commands/doctor_test.go`; preserve meaningful failure assertions. (FR-001, FR-003; depends on T038)
- [ ] T002 On Apple hardware capture Apple container 1.4.1 candidate CLI version/help/inspect/run/network/volume fixtures into `infra/applecontainer/testdata/`; compare real JSON and flags with manager parser; document tested matrix in `docs/runtime-contract.md`. (FR-001, FR-003, FR-012; depends on T001)
- [ ] T003 Execute NET-01–04 for the candidate topology and allowed/denied connectivity matrix in contracts/project-runtime-contract.md, including IPv6/host aliases and shared-gateway network changes; qualify endpoint and DNS strategy in `infra/applecontainer/`, `platform/dns/`, and `docs/runtime-contract.md`; fail gate if bare-name/default-network assumptions remain. (FR-003, FR-007, FR-013, FR-014; depends on T002)
- [ ] T004 Freeze semantic actions, exit codes/JSON schemas, interactive-text semantics and config-origin fixtures including application DB fallback in `core/guidance/`, `cmd/stage/commands/` and `core/config/`; cover no-TTY, init/dry-run/force consent, attach/detach/down/data-deletion transitions, planned --confirm-project and explicit runtime rejection. Freeze inputs/results before T024. (FR-001, FR-002, FR-005, FR-010; depends on T001)
- [x] T005 Add legacy/unknown/newer state, durable-ledger and owner mismatch regression fixtures in `core/state/`; document non-adoption and backup/migration rules in `docs/migration.md`. (FR-004, FR-008, FR-012; depends on T001)
- [ ] T006 Run focused suites, full short suite, vet/build and configured lint; record G0 results and runtime/topology decisions in `specs/012-apple-only-experience/evidence.md` (create). (FR-001, FR-003, FR-012, FR-015; depends on T002–T005, T040)

## Phase 2 — US1: Run my first site (P1, M1/G1)

Independent proof: one routed PHP/database site, stopped and restarted with persistent data. Does not require the finished dashboard.

- [ ] T007 [US1] Add failing adapter/orchestrator integration cases using real captured CLI fixtures for service endpoints, mounts, image failures and total health deadlines in `infra/applecontainer/manager_test.go` and `core/lifecycle/orchestrator_test.go`. (FR-003, FR-013; depends on T006)
- [ ] T008 [US1] Implement proven endpoint/network wiring in `infra/applecontainer/manager.go`, `core/lifecycle/orchestrator.go`, `stacks/20i/apple-container*.json` and `docker/nginx.conf.tmpl`; validate image architecture and PHP/database connectivity. (FR-003, FR-013; depends on T007)
- [ ] T009 [US1] Fix total-deadline health and fresh-process manifest loading in `infra/applecontainer/`; implement APP-01–04 read-only StageServe-owned PHP/DB/route checks before infrastructure success; optional application 500/redirect must not roll back healthy infrastructure in `core/lifecycle/`. (FR-003, FR-008; depends on T008)
- [ ] T010 [US1] Complete Apple readiness and previewed project settings in `core/onboarding/`, `core/config/`, `cmd/stage/commands/init.go` and setup/doctor paths; preserve effective config origins and existing suffixes. (FR-001, FR-002, FR-013; depends on T004, T008)
- [ ] T011 [US1] Run APP-01–04 against a fixture and representative app plus 20-cycle stop/start inventory; test infrastructure-ready with broken user app and no source/readiness DB writes; record G1 evidence and update `docs/runtime-contract.md` and `docs/installer-onboarding.md`. (FR-003, FR-004, FR-013; depends on T009, T010, T041, T042)

## Phase 3 — US2: Operate projects (P1, M2)

Independent proof: two distinct sites and DB sentinels with correctly scoped direct commands; complete UI not required.

- [ ] T012 [US2] Add exact-label ownership, prefix collision and custom-volume regression tests in `infra/applecontainer/manager_test.go` and `core/state/`. (FR-004, FR-007; depends on T011)
- [ ] T013 [US2] Use exact observed ownership and manifest volume identities for stop/delete in `infra/applecontainer/manager.go`; preserve unknown resources and implement/test the specified stop versus unregister versus confirmed volume-deletion effects, including ID-01–04 fresh-process reattach using the independent retained ledger. (FR-004, FR-007, FR-009; depends on T012, T040)
- [ ] T014 [US2] Implement per-route TLS policy and stable installation listeners; reconcile scoped status and shared routing after address changes in `observability/status/`, `core/state/`, `infra/gateway/` and `core/lifecycle/`; define lock order and test concurrent CLI actions. (FR-005, FR-007, FR-008; depends on T013)
- [ ] T015 [US2] Add bounded cancellable log/exec tests and implementation in `observability/logs/`, `infra/applecontainer/` and `cmd/stage/commands/logs.go`; preserve named-service errors and exit results. (FR-006, FR-008, FR-010; depends on T014)
- [ ] T016 [US2] Prove scoped two-project lifecycle/routing/DB isolation and actual local port exposure; execute NET-01–04 negative connectivity and TLS-01–04 mixed HTTP/HTTPS in both operation orders, certificate renewal/refusal/failure, optional debug and host-alias exposure via `platform/dns/`, `platform/tls/` and `stacks/20i/`; record evidence. (FR-006, FR-007, FR-013, FR-014; depends on T015)

## Phase 4 — US3: Recover safely (P1, M2/G2)

Independent proof: failure and cancel matrix preserves data and unrelated projects, using direct operations.

- [ ] T017 [US3] Add phase-by-phase failure/cancel/state-save/route-reload and stale-preview tests in `core/lifecycle/orchestrator_test.go` and `core/guidance/`. (FR-008, FR-009; depends on T016)
- [ ] T018 [US3] Complete cleanup with fresh bounded contexts, leftover-resource reporting and truthful commits in `infra/applecontainer/manager.go` and `core/lifecycle/orchestrator.go`. (FR-004, FR-008; depends on T017)
- [ ] T019 [US3] Implement journal recovery, retained-ledger fresh-process reconciliation and legacy-state refusal in `core/state/`, `observability/status/` and `core/guidance/context.go`; verify operation scope again under lock; test moved-folder refusal, no duplicate registration and retained data, and document manual recovery. (FR-004, FR-008, FR-009, FR-012; depends on T005, T018)
- [ ] T020 [US3] Reconcile destructive and host-change previews/consent in `core/guidance/`, `cmd/stage/commands/` and `platform/dns/`; redact diagnostics and generated sensitive outputs; implement matching --confirm-project for noninteractive down --volumes and reject --all --volumes. (FR-009, FR-014; depends on T019)
- [ ] T021 [US3] Execute Apple failure, runtime-restart, cancellation, ownership and two-project matrix; record G2 in `specs/012-apple-only-experience/evidence.md`; publish recovery instructions in `docs/migration.md` and runtime docs. (FR-004, FR-007, FR-008, FR-009, FR-014; depends on T018–T020)

## Phase 5 — Shared US1/US2/US3 interface completion (M3/G3)

T022–T025 can start after G0 using fixtures while M1/M2 proceed with separate ownership. T026–T030 require real operations and do not complete on fixtures alone.

- [ ] T022 [P] [US1] Reconcile Apple-only examples and default-action/suffix language in `docs/design/`, `mockups/screens.js` and `.github/instructions/terminal-*.instructions.md`; retain existing design work and record rendered review. (FR-001, FR-005, FR-013, FR-015; depends on T006)
- [ ] T023 [P] [US2] Create situation/selected-project/NO_COLOR/text/JSON semantic acceptance fixtures in `core/guidance/` and command tests; include stopped, unknown and stale preview cases. (FR-005, FR-010, FR-011; depends on T006)
- [ ] T024 [US2] After GUX passes, evolve existing `core/guidance/shell.go` into overview/selected-project dashboard and utility views; implement state/facts/action hierarchy, running Enter logs and explicit browser action. (FR-005, FR-006; depends on T022, T023, T039)
- [ ] T025 [US1] Complete continuous readiness->settings preview->run flow using shared actions in `cmd/stage/commands/tui.go`, `core/guidance/` and `core/onboarding/`; implement CFG-01–04 safe existing-project settings apply with no duplicated domain logic. (FR-002, FR-005, FR-006; depends on T024, T041)
- [ ] T026 [US2] Wire measured 5,000-line/2 MiB bounded logs, height-aware viewport/resize, async progress and refresh per ux-validation.md to real operations in `core/guidance/shell.go` and `cmd/stage/commands/tui.go`. (FR-006, FR-010, FR-011; depends on T015, T021, T025)
- [ ] T027 [US3] Wire recovery, cancellation, stale-action revalidation and Cancel-default scoped confirmations through shared operation services; test shell exit while projects run in `core/guidance/` and command integration tests. (FR-008, FR-009, FR-011; depends on T020, T026)
- [ ] T028 [US2] Run A11Y-01–06 with VoiceOver and real 52x24/80x24 keyboard, light/dark, NO_COLOR and resize scenarios; measure p95 navigation, cancellation and log backpressure limits from ux-validation.md; verify redirected text/help/JSON modes and fix defects in their owning modules. (FR-005, FR-006, FR-010, FR-011; depends on T027, T043)
- [ ] T029 [US1] Run the final three representative first-use trials beginning at desktop discovery, including accessible text use and record timing/friction plus browser/DB evidence; resolve critical usability defects before G3 in `specs/012-apple-only-experience/evidence.md`. (FR-002, FR-003, FR-005, FR-011; depends on T028)
- [ ] T030 [US3] Publish G3 evidence and align `README.md`, `docs/design/`, scoped terminal instructions and `mockups/README.md` with verified behaviour. (FR-008, FR-010, FR-011, FR-015; depends on T029)

## Phase 6 — US4: Install and retain work (P2, M4/G4)

Independent proof: verified artifacts install/update/rollback with an existing project; depends on working runtime and UI for end-to-end acceptance.

- [ ] T031 [US4] Implement the release-qualification.md compatibility record: qualified N/N+1 state read/write ranges, pinned image/engine digests, post-write rollback and explicit new-volume migration in `docs/migration.md`, `docs/installer-onboarding.md` and `.github/workflows/`; do not introduce a new updater dependency. (FR-004, FR-012; depends on T006)
- [ ] T032 [US4] Implement REL-01–04/UPD-01–04 using actual candidate artifacts through the production download path, plus clean-install, checksum failure, interrupted-update and rollback smoke cases in `scripts/tests/`; extend `install.sh` and existing release workflow for compatible Apple-only artifacts. (FR-001, FR-004, FR-012; depends on T030, T031, T045)
- [ ] T033 [US4] Rehearse UPD-01–04 including new writes before rollback, incompatible state/engine refusal and quiesced logical DB export/import into a new volume; include legacy config/DB copied fixtures; verify counts/sentinels and restoration; reject automatic legacy adoption in `core/state/` and migration tools/docs. (FR-004, FR-009, FR-012; depends on T019, T032)
- [ ] T034 [US4] Retire unused Docker/Compose execution packages, SDK dependencies and obsolete runtime assets after replacement coverage; retain required OCI build recipes. Audit `go.mod`, `infra/`, lifecycle messages, installer and help for unsupported Docker fallback. (FR-001, FR-012; depends on T021, T032, T033)
- [ ] T035 [US4] Recheck already-corrected planning authority and complete verified operator/help parity in `README.md`, `docs/architecture.md`, runtime/onboarding/migration docs and agent guidance; repair/remove missing historical prototype Makefile target with an explicit supported replacement. (FR-001, FR-010, FR-012, FR-015; depends on T030, T034)
- [ ] T036 [US4] Qualify the exact REL candidate digests; run full short tests, race/integration suites, vet/build/lint, real installed-artifact checks and declared macOS/CLI matrix; record all G4 results and remaining supported limits in evidence and release notes. (FR-001, FR-003, FR-004, FR-007, FR-012, FR-014; depends on T033–T035, T042, T044)
- [ ] T037 [US4] Run scripts/verify-planning.py for 88 content hashes, links and coverage, then final Spec Kit analysis; promote the same digests only when G0, GUX and G1–G4 evidence is complete. (FR-015; depends on T036)

## Review repairs and additional obligations

IDs are stable; dependencies, not numeric order, determine execution. T038 is the
completed planning repair; all research/production tasks below remain unchecked.

- [x] T038 Correct current/archive authority, preserve 88 content files plus original 90-file local metadata provenance, and add scripts/verify-planning.py. (FR-015; depends on none)
- [ ] T039 [P] [US1] Execute UX-01–05 early five-participant study in ux-validation.md using fixture revision and desktop discovery; publish GUX decision in evidence.md before T024. Failed terminal-entry thresholds require revised surface decision and estimate; do not simulate participants. (FR-022; depends on T004)
- [ ] T040 [US3] Implement immutable UUID identities/<uuid>.json retained ledger, exact resource labels and write-ahead journal in core/state/, core/lifecycle/ and infra/applecontainer/; implement ID-01–04 including detach/copy/missing-ledger/crash recovery tests. (FR-004, FR-009, FR-017; depends on T005)
- [ ] T041 [US2] Implement CFG-01–04 settings preview/apply transaction in core/config/, core/lifecycle/, core/state/ and command adapters, keyed desired/applied fingerprint, field policy, rollback and stopped apply; preserve app files and DB contents. (FR-002, FR-016, FR-018; depends on T008, T040)
- [ ] T042 [US2] Implement and measure OFF-01–03 and CAP-01–03 in infra/applecontainer/, manifests and observability/ on minimum host: exact cached image reuse, missing-cache refusal before downtime, CPU/RAM/disk limits, ENOSPC recovery and bounded logs; record manifest defaults and G1 re-estimate. (FR-020, FR-021; depends on T009)
- [ ] T043 [US1] Implement the complete interactive line-oriented --cli/--notui journey and A11Y-01–06 in core/guidance/ and command adapters; preserve non-TTY/JSON semantics; validate actual VoiceOver use and shared action parity. (FR-010, FR-011, FR-023; depends on T025, T039)
- [ ] T044 [US4] Implement/test HOST-01–03 sleep/wake/reboot reconciliation and supervised uninstall/reinstall procedure in core/lifecycle/, platform/, install.sh and docs/installer-onboarding.md; preserve retained data, externally edited resolver files and unrelated Apple resources. (FR-004, FR-009, FR-021; depends on T021, T031)
- [ ] T045 [US4] Replace tag-immediate-publish in .github/workflows/release.yml with build-candidate/qualify/promote dependencies, actual binary/bundle/installer/checksum inventory and same-digest promotion in contracts/release-qualification.md; block publication for absent evidence, mismatch or unqualified platform. Test negative promotion cases without publishing. (FR-012, FR-019; depends on T030, T031)

## Dependencies, acceptance and estimates

Execution order: T038 -> T001 -> T002/T004/T005; T005 -> T040;
T002 -> T003; T003/T004/T005/T040 -> T006; T006 -> T007 -> T008 -> T009;
T008/T040 -> T041 and T009 -> T042; T010/T009/T041/T042 -> T011;
then T012–T021. T039 follows T004 and gates T024 independently of live runtime.
T022/T023 fixtures follow G0; T024 -> T025 -> T043, while T026 waits for G2.
T028 waits for both T027 and T043. T031 begins after G0; T045 follows T030/T031,
then T032/T033/T034/T035. T044 follows T021/T031. T036 joins all qualification
work; T037 checks planning and permits same-digest promotion after all gates.
No resource-sharing edits are parallel without exact file ownership.

Estimates in docs/roadmap.md replace the prior 6–12.5-week allowance. All SCs have
owned cases: SC001 T029; SC002 T011; SC003 T021/T027; SC004 T023/T028/T043;
SC005 T016/T021; SC006 T032/T033/T036; SC007 T038/T037; SC008 T040/T041;
SC009 T032/T033/T045; SC010 T042/T044; SC011 T039/T028/T043.

## Requirement coverage

| Requirement | Implementation tasks | Verification tasks |
|---|---|---|
| FR-001 | T002, T004, T010, T034 | T006, T036 |
| FR-002 | T010, T025 | T004, T029 |
| FR-003 | T008, T009 | T007, T011, T036 |
| FR-004 | T013, T018, T019, T032 | T005, T011, T021, T033 |
| FR-005 | T004, T024, T025 | T023, T028, T029 |
| FR-006 | T015, T024, T026 | T016, T028 |
| FR-007 | T003, T013, T014 | T012, T016, T021 |
| FR-008 | T009, T018, T019, T027 | T017, T021 |
| FR-009 | T013, T019, T020, T027 | T017, T021, T033 |
| FR-010 | T004, T015, T026 | T023, T028, T035 |
| FR-011 | T026, T027 | T023, T028, T029 |
| FR-012 | T002, T019, T031, T032, T034 | T005, T033, T036 |
| FR-013 | T003, T008, T010 | T011, T016 |
| FR-014 | T020 | T016, T021, T036 |
| FR-015 | T022, T030, T035 | T037 |
| FR-016 | T041, T025 | T041, T028 |
| FR-017 | T040, T013, T019 | T040, T021 |
| FR-018 | T004, T009, T041 | T011, T041 |
| FR-019 | T031, T032, T045 | T033, T036, T045 |
| FR-020 | T042 | T042, T036 |
| FR-021 | T042, T026, T044 | T042, T028, T044, T036 |
| FR-022 | T039 | T039, T029 |
| FR-023 | T043 | T028, T043 |
