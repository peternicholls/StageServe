# Research and decision record

Revised: 2026-09-13 (initial review 2026-09-11). Scope: current checkout, old specs 002–011, design guides and mockup source. These decisions resolve planning choices; they do not substitute for runtime experiments.

## Source-backed baseline

| Area | Evidence | Consequence |
|---|---|---|
| Apple-only selection | `core/runtime/types.go:15–30`; `cmd/stage/commands/root.go:136–148`; `cmd/stage/commands/tui.go:61–98` | Continue current wiring; do not restart a hybrid adapter project |
| Guided UI | `core/guidance/planner.go`; `core/guidance/shell.go`; `cmd/stage/commands/tui.go` | Existing production shell is the implementation starting point |
| Service addressing | `infra/applecontainer/manager.go:238–242` uses default network; `docker/nginx.conf.tmpl:40` uses `apache:9000`; adapter says BareNetworkDNS false | Resolve and prove endpoint injection; current assets are not live-compatible evidence |
| Ownership | `infra/applecontainer/manager.go:351–368` prefix filtering, despite labels on create | Require exact ownership and collisions tests before deletion |
| Rollback/health | `infra/applecontainer/manager.go:109–119,208–225,295–311,344–348` | Cover fresh-process health, global deadlines, cancelled-context cleanup and partial starts |
| Docker remnants | `infra/docker`, `infra/compose`, `go.mod`; lifecycle/shell remediation strings | Remove runtime dependencies and obsolete advice after contract-preserving replacements; OCI Dockerfiles can remain |
| Interface gaps | `core/guidance/shell.go:49–69,125–128`; `cmd/stage/commands/tui.go:74–84,176–189` | Project overview, height-aware viewport and bounded live logs need explicit tasks |

Review host: macOS 27.0, arm64, Go 1.27.1. Apple `container` absent from PATH. `go.mod` still requests Go 1.26 / toolchain 1.26.2; local test success under a newer toolchain is not a release-matrix result.

Initial focused baseline results (before concurrent commits): `go test ./core/runtime ./infra/applecontainer ./infra/runtime`, `./observability/status`, and `./core/guidance ./core/onboarding ./core/config ./core/state` passed. Lifecycle failed `TestOrchestrator_AttachAddsRouteAndMarksAttached` (old `stage-demo-web` versus observed `stage-demo-nginx.test`). Commands failed `TestDoctor_TextOutputShape` (expects Docker). `observability/logs` has no tests. No runtime services, DNS or trust settings were changed.

## Decisions and alternatives

### D1 — TUI first, shared core

Decision: guided TUI plus direct CLI/text/JSON, not a separate v1 GUI. Reuse the existing Go/Charm implementation and core planner.
Rationale: `docs/concept.md`, design guides, 007 and 011 agree on an assisted terminal journey; mockups cover its states. Alternatives: native GUI improves discovery but adds packaging, accessibility and transport work; browser GUI introduces server/security surface without a demonstrated need. Validate discovery, terminal entry and accessibility at GUX before full M3 implementation; failure thresholds require a decision amendment.

### D2 — Apple runtime only

Decision: one supported runtime; reject all other selector values. Preserve domain interfaces where useful for tests, not for hypothetical backend plugins. OCI images/build recipes are not Docker runtime dependencies.
Rationale: explicit user direction and current ParseBackend. Alternative hybrid compatibility conflicts with scope and doubles acceptance surfaces.

### D3 — Explicit endpoints and real topology proof

Decision: qualify the private application/ingress network candidate and explicit connectivity matrix in contracts/project-runtime-contract.md. Runtime observations identify per-interface addresses; app/gateway configuration is reconciled after changes. G0 must prove the matrix, including negative paths and safe gateway network changes, before accepting the candidate.
Rationale: current name assumptions contradict declared capabilities. Alternative blindly copying Compose networking cannot pass G1. The topology is an explicit candidate with a rejection rule; default-network use cannot satisfy the private-service policy. If Apple multi-network behaviour fails it, an architecture decision must qualify an alternative without weakening isolation.

### D4 — Preserve identity and data

Decision: retain atomic JSON persistence with an independent identities/<uuid>.json retained-data ledger and apply journal; registry membership is a projection, so detach does not erase identity or volume ownership. Missing/legacy runtime identity remains legacy until intentional migration. Owned labels plus validated identity, not loose prefixes, authorize mutations. Keep existing data untouched during planning.
Rationale: older documents normalize missing runtime as Docker; switching that default would falsely adopt resources. Alternative implicit adoption is unsafe. DB migration is backed-up logical export/import with verification, not volume renaming.

### D5 — One canonical interaction contract

Decision: dashboard hierarchy from the current proposal; running Enter views logs; stopped Enter runs; destructive modal defaults Cancel. Explicit browser action remains available. Default new suffix `.test`; opt-in TLS displays actual scheme/port. Keep setup/doctor read-only direct reports, with optional guided assistance as a separate action.
Rationale: resolves default-action and `.test`/`.develop`/`.dev` inconsistencies across mockups and docs. Commands and projections share semantic fixtures.

### D6 — Packaging and migration before release

Decision: compatible binary plus versioned runtime assets, verified checksum/install/update rollback tests, explicit preservation of user config/state. Keep current install path compatibility initially; do not add another rename migration. Retire Compose execution/SDK dependencies by G4 while retaining needed OCI image build inputs.
Alternative: embedding assets may later reduce drift, but requires an independent benefit check against existing runtime-bundle work.

## Design and mockup review

Read: `docs/design/README.md`, visual identity, copy and experience guides, dashboard proposal, component prototypes, guided flow map; `docs/superpowers/specs/2026-05-07-guided-tui-prototype-design.md`; `mockups/index.html`, `app.js`, `screens.js`, `styles.css`.

Retain: verdict-first header, compact evidence, local commands, defaults preview, calm error copy, semantic colour, More for advanced tools, utility logs/status, fixture-only design viewer. `screens.js` covers ready/running/loading, both confirmation classes, missing prerequisites, editor, not-project/stopped/mismatch, details, doctor, More/palette, text/JSON, narrow/no-colour, init/logs.

Correct before shipping: Docker prerequisite/remediation text; duplicate dashboard/body facts; no guaranteed first-level implementation jargon; `.test` versus `.develop` examples; actual Enter action; colour-independent markers; narrow layout and terminal height; clear selected project; logs viewport and stale-state/cancel screens. CSS terminal chrome and gradients are illustrative, not acceptance requirements.

Review limitation: local-file navigation was blocked by browser policy. Source inspection is complete; no rendered browser or real-terminal visual pass is claimed. M3 owns that acceptance. `Makefile` points to a missing `specs/007.../prototype` directory; do not pretend the historical prototype is runnable. The maintained runnable design reference is `docs/design/style-guide-tui`.

## External evidence and version limits

Apple's [repository](https://github.com/apple/container) describes Apple silicon and macOS 26 support. Its [technical overview](https://github.com/apple/container/blob/main/docs/technical-overview.md) explains per-container VMs and the runtime service. Its [command reference](https://github.com/apple/container/blob/main/docs/command-reference.md) is the starting point for CLI integration. Reviewed 2026-09-11. These are main-branch documents, not a pinned compatibility guarantee: M0 must capture the installed release's help and real output before approving adapter fixtures.

No new external libraries are required by this plan. Networking, supported version range and hardware performance remain measured milestone outputs; product-scope choices have normative contracts; empirical claims remain open until the named experiments pass.

## Refreshed baseline after concurrent work

During planning the shared checkout advanced through `8d537ef` (runtime work), `e1eb881` (contract/design work, including some planning files) and `3998a3b` (style guide). This planning task did not create those commits and does not attribute their source edits to this review. Existing work was not reverted.

On the refreshed `3998a3b` checkout plus planning edits, `go test -short ./...`, `go vet ./...` and `go build ./...` passed. The two initially failing tests now pass. `make lint` fails its prerequisite check because `staticcheck` is not installed; no dependency was installed. `container` is still unavailable, so no live runtime or rendered visual acceptance is claimed. Initial line references above describe the inspected snapshot and may shift; recheck the relevant symbols before implementation.

A read-only review of the related task "Review proposed direction change" confirmed the earlier hybrid migration was interrupted rather than accepted as complete. Current user direction and spec 012 supersede those hybrid assumptions.

Draft review resolved four issues: exact attach/detach/down/data-deletion transitions; moved-folder refusal instead of a promised relocation feature; explicit init/force/dry-run automation consent; and freezing action contracts before parallel UI work. See application-contract.md and T004/T013/T019/T020.

## Research repair: versioned primary evidence

Reviewed 2026-09-13. Candidate is **Apple container 1.4.1**, not a claim that it is
installed or supported by StageServe. The [1.4.1 release](https://github.com/apple/container/releases/tag/1.4.1)
includes changed system-status reporting, so old readiness fixtures cannot establish
compatibility. Choose this released candidate instead of an unspecified main branch;
recheck support/security advisories at qualification time.

[Versioned networking documentation](https://github.com/apple/container/blob/1.4.1/docs/networking.md)
describes isolated networks, per-network addresses, loopback publishing, interface
order for published ports and limitations of bare-name resolution. This supports
using inspected addresses and testing multi-network/host-alias paths; it does not
prove our proposed gateway topology on this host.

[Versioned volumes documentation](https://github.com/apple/container/blob/1.4.1/docs/volumes.md)
describes named-volume inspection and sparse capacity. Capacity is not actual disk
consumption; the capacity experiment must measure both. [Command reference](https://github.com/apple/container/blob/1.4.1/docs/command-reference.md)
is the source for exact argv, label and inspect-shape fixtures. No global prune is
part of StageServe cleanup. Primary documents guide experiments; passing results
must come from the candidate CLI on actual supported hardware.

## Research register: execute before the dependent decision

| Research ID / owner task | Inputs and experiment | Required recorded output | Passing decision / failure response |
|---|---|---|---|
| R-01 / T002 | 1.4.1 on macOS 26 and 27 arm64; version/help/list/inspect/volume/network/system status | CLI fixtures, exact builds, command exit/error schemas and version support rows | Accept only passing rows; reject unknown versions; update parser before G0 |
| R-02 / T003 | Two disposable projects on application/ingress candidate, gateway interface changes, IPv4/v6/host aliases | NET-01–04 complete allowed/denied results and diagram with observed interfaces | All forbidden private-service paths fail and existing site survives topology change; otherwise block G0 and record new Apple-only architecture decision |
| R-03 / T003 | Existing dnsmasq provider versus Apple DNS for aliases, .test/.develop/.dev, resolver conflict and unrelated Apple workloads | Setup effects, owner hashes, repeatability, rollback and any required service restart | Qualify dnsmasq candidate; change provider only if alias/suffix/ownership requirements pass with less setup and no unrelated-workload restart |
| R-04 / T009/T011 | StageServe PHP probe + SELECT 1; empty app, redirect, PHP 500 and app DB config mismatch | APP-01–04 responses and source/database before-after hashes/sentinels | Infrastructure proof without application mutation; bad app stays inspectable and running |
| R-05 / T040/T041 | Detach/copy/rename, journal fault points, stopped/running settings changes, failed recreation | ID/CFG exact retained identity, desired/applied hashes, leftovers and rollback results | No false applied status or data adoption/loss; otherwise block G1 |
| R-06 / T042 | Minimum 16 GiB host, two projects, disconnected cached start/recreation, disk-full and log load | OFF/CAP timings, memory/CPU/disk, selected manifest resource defaults and missing-cache output | Meet proposed bounds or amend minimum/limits with evidence before G1; no silent target relaxation |
| R-07 / T039 | Five participants incl low terminal familiarity and accessible user; desktop discovery tasks | Anonymised UX-01–05 observations and threshold decision | GUX passes before T024 or revise onboarding/interface hypothesis and estimate |
| R-08 / T043/T028 | Actual VoiceOver/text/keyboard journeys and slow-runtime/log fixtures | A11Y-01–06 and p95/cancel/backpressure metrics | Complete safe text journey plus measured UI limits before G3 |
| R-09 / T031/T032/T045 | N/N+1 actual artifacts, state/engine incompatibility, new DB writes before rollback | REL/UPD digest-qualified install/rollback and negative promotion evidence | Same digests, new writes preserved, incompatible paths refused; otherwise block G4 |
| R-10 / T044 | Sleep/reboot/runtime restart; uninstall/reinstall with externally modified resolver and unrelated workload | HOST-01–03 observations and restore procedure | No automatic resume or unrelated mutation; retained data recoverable before G4 |

All R-01–R-10 empirical results are **pending**, not completed by this planning repair.
No participant, benchmark, runtime or signing evidence is invented. The repair closes
missing definitions, sources, owners and decision thresholds; task execution supplies
actual results. No release or full-product readiness is implied.
