# Planning handoff and acceptance runbook

This is a future acceptance procedure. It is not evidence that Apple runtime acceptance has passed.

## Resume Spec Kit

From repository root:

```sh
export SPECIFY_FEATURE=012-apple-only-experience
bash .specify/scripts/bash/check-prerequisites.sh --json --require-tasks --include-tasks
```

The Git branch uses `codex/`; `SPECIFY_FEATURE` supplies the numeric name expected by the current Spec Kit scripts. Do not rerun setup-plan on completed documents: it copies the template over plan.md. Start implementation with the first unchecked task only after reading the constitution, research and current working-tree diff.

## Baseline checks (no runtime mutation)

```sh
go test ./core/runtime ./infra/applecontainer ./infra/runtime
go test ./core/guidance ./core/onboarding ./core/config ./core/state
go test ./core/lifecycle ./cmd/stage/commands ./observability/status
go test -short ./...
go vet ./...
go build ./...
git diff --check
```

Use Makefile lint once its configured tool is available; report any unavailable gate rather than installing new dependencies implicitly. M0 captures the exact toolchain and CLI help/inspect fixtures. Baseline failures are recorded in research.md and must be reconciled before G0 closes.

## Live acceptance matrix

Use disposable explicitly named fixture projects and a separate StageServe test state directory. Inventory owned resources first; never use global prune. Record OS/arch, Apple CLI version, StageServe revision, manifest/image digests, configuration and outputs. Do not install dependencies or change DNS/trust as a side effect of merely reviewing this runbook.

| Case | Steps | Pass condition | Gate |
|---|---|---|---|
| Host/CLI contract | Check supported and unsupported host/runtime; capture version/help/inspect output | Precise fail-before-mutation for unsupported combination; real adapter fixtures match | G0 |
| One site | Init with preview, up, visit displayed URL; PHP reads/writes DB sentinel | Browser response and DB roundtrip prove full path, StageServe-owned infrastructure checks precede success; application errors remain separate | G1 |
| Persistence/repeatability | Down/up 20 times; inventory resources each cycle | Same sentinel, identity and hostname; no accumulated containers/routes/env files | G1 |
| Two projects | Unique content/DB sentinel per project; stop/restart one, repair shared route | No response/data cross-over or damage to other project | G2 |
| Failure per startup phase | Fail image/build, mount, DB health, web health, route reload, state save | No false attached record; owned cleanup/leftovers reported; unrelated resources unchanged | G2 |
| Cancellation/restart | Cancel each phase; restart StageServe and runtime service | Bounded cleanup independent of cancelled context; fresh-process status accurate | G2 |
| Ownership/conflict | Similar prefixes, mismatched labels, duplicate names/ports, stale/newer records | Reject/diagnose, never adopt/delete ambiguous or unrelated resources | G2 |
| Command transitions | Exercise down, detach, attach, confirmed volume deletion, rejected all-volume deletion and moved-folder refusal | Exact route/record/service/data effects match contract; no duplicate registration | G2 |
| Init automation | Run interactive cancel, dry-run/JSON, direct noninteractive defaults and force overwrite on fixtures | No writes on cancel/dry-run/invalid input; existing file retained without force | G0/G3 |
| Local DNS/TLS | TLS-01–04 mixed HTTP/HTTPS in both operation orders; negative connectivity NET-01–04 | Exact preview, per-route policy, no forbidden private/host-alias/IPv6 connection or unrelated route change | G2 |
| Logs/exec | Missing service, large stream, cancellation, failure exit | Bounded viewport/buffer, usable exit, preserved exit status, no secrets in reports | G2/G3 |
| Terminal parity | All planner situations at 52x24 and 80x24, resize, light/dark, NO_COLOR | Scope/actions visible, no trapping, semantic parity with text, keyboard-only completion | G3 |
| Automation | Redirect stdin/stdout, no-TUI aliases, JSON/help | No prompt/ANSI/UI pollution; stable structured contract and exit semantics | G3 |
| Usability | Three representative final trials beginning at desktop discovery, including accessible text | Complete session without undocumented ordering, <=10 min excluding downloads target; critical blockers resolved | G3 |
| Distribution | REL/UPD cases using exact candidate through production downloads; update then new DB writes then rollback | Same digest promotion; backwards-readable state and latest data preserved; incompatible rollback refused | G4 |
| Legacy migration | Copy/inventory Docker-era records and DB export; import into isolated Apple fixture | Source retained; sentinel/count checks and restore steps pass; no implicit adoption | G4 |

## Additional mandatory cases from review

| Case family | Procedure source | Task / gate |
|---|---|---|
| ID-01–04 | [Durable ledger and detach/copy/crash](contracts/project-runtime-contract.md) | T040/T013/T019, G0/G2 |
| CFG-01–04 | [Settings transaction and field effects](contracts/project-runtime-contract.md) | T041/T025, G1/G3 |
| APP-01–04 | [Read-only PHP/DB and bad application cases](contracts/project-runtime-contract.md) | T009/T011, G1 |
| NET-01–04 / TLS-01–04 | [Connectivity matrix and mixed TLS](contracts/project-runtime-contract.md) | T003/T016, G0/G2 |
| UX-01–05 | [Five-participant early discovery study](ux-validation.md) | T039, GUX before T024 |
| A11Y-01–06 | [VoiceOver interactive text and TUI parity](ux-validation.md) | T043/T028, G3 |
| OFF-01–03 / CAP-01–03 | [Disconnected recreation, minimum host and exhaustion](contracts/release-qualification.md) | T042, G1 |
| HOST-01–03 | [Sleep/reboot/uninstall](contracts/release-qualification.md) | T044, G4 |
| REL-01–04 / UPD-01–04 | [Exact artifacts and post-write rollback](contracts/release-qualification.md) | T032/T033/T045, G4 |

Run `python3 scripts/verify-planning.py` for planning links, tasks/dependencies,
requirement coverage, instruction authority and the 88-file clean-checkout archive
inventory. The original 90-file snapshot includes optional local Finder metadata.
Planning checks cannot close any runtime, participant or release qualification case.

## Evidence format

For each case record date, revision/worktree diff, host/CLI/toolchain, fixture identity, exact command or interaction, expected/actual result, evidence path and unresolved gap. Screenshots support visual checks; HTTP/DB assertions support runtime claims. Keep credentials and personal project paths out of shared evidence. A skipped case remains open.

## Current verification limits

Initial review found focused passes and two failing suites; after concurrent changes, full short tests, vet and build pass. `make lint` is blocked because staticcheck is absent. See research.md for the refreshed baseline. No `container` executable is available, no live application path was exercised, and local browser mockup navigation was policy-blocked. Design source review is not visual acceptance. Planning completion closes none of G0/GUX/G1–G4.
